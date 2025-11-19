package live

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"

	"github.com/fanchann/docunyan/internals/constants"
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

	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		return errors.New(string(out))
	}
	return nil
}

//go:embed index.html
var index string

var upgrader = websocket.Upgrader{
	ReadBufferSize:  constants.WebSocketReadBufferSize,
	WriteBufferSize: constants.WebSocketWriteBufferSize,
}

func SwaggerLive(fileName string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		cancel()
	}()

	if err := swaggerLiveWithContext(ctx, fileName); err != nil {
		log.Fatal(err)
	}
}

func swaggerLiveWithContext(ctx context.Context, fileName string) error {
	msg := make(chan []byte)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	fi, err := os.Stat(fileName)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	old := fi.ModTime()

	err = watcher.Add(filepath.Dir(fileName))
	if err != nil {
		return fmt.Errorf("failed to add watcher: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if filepath.Base(event.Name) == filepath.Base(fileName) {
					time.Sleep(constants.FileWatcherDebounce)
					fi, err := os.Stat(fileName)
					if err != nil {
						log.Println(err)
						continue
					}
					now := fi.ModTime()
					if !old.Equal(now) {
						old = now
						log.Println("update")
						b, err := os.ReadFile(fileName)
						if err != nil {
							log.Println(err)
							continue
						}
						msg <- b
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer c.Close()

		b, err := os.ReadFile(fileName)
		if err != nil {
			log.Println(err)
			return
		}

		resp := map[string]interface{}{
			"fileName": fileName,
			"message":  string(b),
		}
		if err := c.WriteJSON(resp); err != nil {
			log.Println(err)
			return
		}

		done := make(chan bool)
		go func() {
			_, _, _ = c.ReadMessage()
			done <- true
		}()

		for {
			select {
			case <-ctx.Done():
				log.Println("context cancelled, closing websocket")
				return
			case m := <-msg:
				if err := c.WriteJSON(map[string]string{"message": string(m)}); err != nil {
					log.Println(err)
					return
				}
			case <-done:
				log.Println("close websocket")
				return
			}
		}
	})

	port, err := utils.GetAvailableRandomPort()
	if err != nil {
		return fmt.Errorf("failed to get available port: %w", err)
	}

	portStr := strconv.Itoa(port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body := fmt.Sprintf(index, portStr)
		_, _ = w.Write([]byte(body))
	})

	server := &http.Server{
		Addr: ":" + portStr,
	}

	// Start server in goroutine
	go func() {
		log.Println("start server:", port)
		log.Println("watching", fileName)

		if err := openBrowser("http://localhost:" + portStr); err != nil {
			log.Println("cannot open browser", err)
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Println("Server stopped gracefully")
	return nil
}
