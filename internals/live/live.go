package live

import (
	_ "embed"
	"errors"
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

	"github.com/fanchann/docunyan/internals/utils"
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

func SwaggerLive(fileName string) {
	msg := make(chan []byte)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer watcher.Close()

	fi, err := os.Stat(fileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	old := fi.ModTime()

	err = watcher.Add(filepath.Dir(fileName))
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if filepath.Base(event.Name) == filepath.Base(fileName) {
					time.Sleep(100 * time.Millisecond) // debounce
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

	var upgrader = websocket.Upgrader{}

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
		panic(err)
	}

	portStr := strconv.Itoa(port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body := fmt.Sprintf(index, portStr)
		_, _ = w.Write([]byte(body))
	})
	log.Println("start server:", port)
	log.Println("watching", fileName)

	if err := openBrowser("http://localhost:" + portStr); err != nil {
		log.Println("cannot open browser", err)
	}
	log.Fatal(http.ListenAndServe(":"+portStr, nil))
}
