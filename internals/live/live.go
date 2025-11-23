package live

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"

	"github.com/fanchann/docunyan/internals/constants"
	"github.com/fanchann/docunyan/internals/errors"
	"github.com/fanchann/docunyan/internals/utils"

	_ "embed"
)

func openBrowser(url string) error {
	args := []string{}
	switch runtime.GOOS {
	case "windows":
		r := strings.NewReplacer("&", "^&")
		args = []string{"cmd", "start", "/", r.Replace(url)}
	case "linux":
		args = []string{"xdg-open", url}
	case "darwin":
		args = []string{"open", url}
	}

	if err := exec.Command(args[0], args[1:]...).Run(); err != nil {
		return fmt.Errorf("failed to open browser: %v", err)
	}
	return nil
}

//go:embed index.html
var index string

var upgrader = websocket.Upgrader{
	ReadBufferSize:  constants.WebSocketReadBufferSize,
	WriteBufferSize: constants.WebSocketWriteBufferSize,
}

// SwaggerLive starts the live preview server with proper lifecycle management
func SwaggerLive(fileName string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return SwaggerLiveWithContext(ctx, fileName)
}

// SwaggerLiveWithContext starts the live preview server with provided context
func SwaggerLiveWithContext(ctx context.Context, fileName string) error {
	msg := make(chan []byte, 10) // Buffered channel to prevent blocking

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.NewServerError("", "failed to create file watcher", err)
	}
	defer watcher.Close()

	// Validate file exists and get initial stats
	fi, err := os.Stat(fileName)
	if err != nil {
		return errors.NewFileSystemError(fileName, "failed to access swagger file", err)
	}
	old := fi.ModTime()

	// Watch the directory containing the swagger file
	watchDir := filepath.Dir(fileName)
	if err := watcher.Add(watchDir); err != nil {
		return errors.NewWatcherError(watchDir, "failed to setup file watching", err)
	}

	// Start file watcher goroutine
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		defer close(msg)
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if filepath.Base(event.Name) == filepath.Base(fileName) {
					// Use debouncing to prevent rapid-fire updates
					select {
					case <-time.After(constants.FileWatcherDebounce):
						if err := checkAndSendFileUpdate(fileName, &old, msg); err != nil {
							log.Printf("Error updating file: %v", err)
						}
					case <-ctx.Done():
						return
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("File watcher error: %v", err)
			}
		}
	}()

	// Get available port
	port, err := utils.GetAvailableRandomPort()
	if err != nil {
		return errors.NewServerError("", "failed to get available port", err)
	}

	portStr := strconv.Itoa(port)

	// Create HTTP server with proper configuration
	server := &http.Server{
		Addr:         ":" + portStr,
		ReadTimeout:  constants.ServerReadTimeout,
		WriteTimeout: constants.ServerWriteTimeout,
	}

	// Setup WebSocket handler with security checks
	upgrader := websocket.Upgrader{
		ReadBufferSize:  constants.WebSocketReadBufferSize,
		WriteBufferSize: constants.WebSocketWriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			// Allow localhost and 127.0.0.1 in development
			return origin == "" ||
				strings.HasPrefix(origin, "http://localhost") ||
				strings.HasPrefix(origin, "http://127.0.0.1")
		},
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}

		if err := handleWebSocket(ctx, c, fileName, msg); err != nil {
			log.Printf("WebSocket handler error: %v", err)
		}
	})

	// Setup main page handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body := fmt.Sprintf(index, portStr)
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(body)); err != nil {
			log.Printf("Error writing response: %v", err)
		}
	})

	// Start server in managed goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Starting live preview server on port %d", port)
		log.Printf("Watching file: %s", fileName)

		if err := openBrowser(fmt.Sprintf("http://localhost:%s", portStr)); err != nil {
			log.Printf("Could not open browser: %v", err)
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		// Context cancelled, perform graceful shutdown
		return shutdownServer(server, 5*time.Second)
	case err := <-serverErr:
		return errors.NewServerError("", "server error", err)
	}
}

// handleWebSocket manages WebSocket connections with proper lifecycle
func handleWebSocket(ctx context.Context, c *websocket.Conn, fileName string, msg <-chan []byte) error {
	defer c.Close()

	// Send initial file content
	b, err := os.ReadFile(fileName)
	if err != nil {
		return errors.NewFileSystemError(fileName, "failed to read swagger file", err)
	}

	resp := map[string]interface{}{
		"fileName": fileName,
		"message":  string(b),
	}

	if err := c.WriteJSON(resp); err != nil {
		return errors.NewServerError("", "failed to send initial data", err)
	}

	// Handle WebSocket communication
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, _, err := c.ReadMessage()
				if err != nil {
					return // Client disconnected
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case m, ok := <-msg:
			if !ok {
				return nil // Channel closed
			}
			if err := c.WriteJSON(map[string]string{"message": string(m)}); err != nil {
				return errors.NewServerError("", "failed to send update", err)
			}
		case <-done:
			return nil // Client disconnected
		}
	}
}

// checkAndSendFileUpdate checks if file has changed and sends update if needed
func checkAndSendFileUpdate(fileName string, oldModTime *time.Time, msg chan<- []byte) error {
	fi, err := os.Stat(fileName)
	if err != nil {
		return err
	}

	now := fi.ModTime()
	if !oldModTime.Equal(now) {
		*oldModTime = now
		log.Println("File updated, broadcasting changes")

		b, err := os.ReadFile(fileName)
		if err != nil {
			return err
		}

		select {
		case msg <- b:
		default:
			// Channel full, skip this update
		}
	}

	return nil
}

// shutdownServer performs graceful server shutdown
func shutdownServer(server *http.Server, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Println("Shutting down server gracefully...")

	if err := server.Shutdown(ctx); err != nil {
		return errors.NewServerError("", "server shutdown error", err)
	}

	log.Println("Server stopped gracefully")
	return nil
}
