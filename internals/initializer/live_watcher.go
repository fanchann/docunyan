package initializer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"

	"github.com/fanchann/docunyan/internals/constants"
	"github.com/fanchann/docunyan/internals/generator"
	"github.com/fanchann/docunyan/internals/live"
)

type LiveWatcher struct {
	config  InitConfig
	watcher *fsnotify.Watcher
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewLiveWatcher(config InitConfig) (*LiveWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &LiveWatcher{
		config:  config,
		watcher: watcher,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

func (lw *LiveWatcher) Start() error {
	defer lw.watcher.Close()

	// Watch config file directory
	configDir := filepath.Dir(lw.config.ConfigPath)
	if configDir == "" || configDir == "." {
		configDir = "."
	}
	if err := lw.watcher.Add(configDir); err != nil {
		return fmt.Errorf("failed to watch config directory: %w", err)
	}

	// Watch DTO directory (can be a directory or file)
	dtoPath := lw.config.DTOPath
	info, err := os.Stat(dtoPath)
	if err != nil {
		return fmt.Errorf("failed to stat DTO path: %w", err)
	}

	var dtoDir string
	if info.IsDir() {
		dtoDir = dtoPath
	} else {
		dtoDir = filepath.Dir(dtoPath)
	}

	if dtoDir != configDir {
		if err := lw.watcher.Add(dtoDir); err != nil {
			return fmt.Errorf("failed to watch DTO directory: %w", err)
		}
	}

	// Initial generation
	color.HiCyan("Docunyan Live Mode")
	color.HiCyan("====================")
	fmt.Println()

	if err := lw.regenerate(); err != nil {
		color.Red("❌ Initial generation failed: %v", err)
		return err
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start live preview in background
	go func() {
		time.Sleep(500 * time.Millisecond) // Wait for initial generation
		if _, err := os.Stat(lw.config.OutputPath); err == nil {
			live.SwaggerLive(lw.config.OutputPath)
		}
	}()

	// Watch for file changes
	debounceTimer := time.NewTimer(0)
	<-debounceTimer.C // drain initial timer

	for {
		select {
		case <-lw.ctx.Done():
			color.Yellow("\n👋 Stopping live watcher...")
			return nil

		case <-sigChan:
			color.Yellow("\n👋 Received interrupt signal, shutting down...")
			lw.cancel()
			return nil

		case event, ok := <-lw.watcher.Events:
			if !ok {
				return fmt.Errorf("watcher event channel closed")
			}

			// Check if it's one of our watched files
			eventFile := filepath.Clean(event.Name)
			configFile := filepath.Clean(lw.config.ConfigPath)

			// Check if event is config file or any .go file in DTO directory
			isConfigFile := eventFile == configFile
			isDTOFile := false

			// Check if it's a .go file in the DTO directory
			if strings.HasSuffix(eventFile, ".go") && !strings.HasSuffix(eventFile, "_test.go") {
				dtoInfo, err := os.Stat(lw.config.DTOPath)
				if err == nil {
					var dtoDir string
					if dtoInfo.IsDir() {
						dtoDir = filepath.Clean(lw.config.DTOPath)
					} else {
						dtoDir = filepath.Clean(filepath.Dir(lw.config.DTOPath))
					}

					eventDir := filepath.Clean(filepath.Dir(eventFile))
					isDTOFile = eventDir == dtoDir
				}
			}

			if isConfigFile || isDTOFile {
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					// Debounce: reset timer
					debounceTimer.Reset(constants.FileWatcherDebounce)

					go func() {
						<-debounceTimer.C
						color.HiYellow("\n📝 File changed: %s", filepath.Base(eventFile))
						color.HiYellow("⏳ Regenerating...")

						if err := lw.regenerate(); err != nil {
							color.Red("❌ Regeneration failed: %v", err)
						}
					}()
				}
			}

		case err, ok := <-lw.watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher error channel closed")
			}
			color.Red("⚠️  Watcher error: %v", err)
		}
	}
}

func (lw *LiveWatcher) regenerate() error {
	start := time.Now()

	// Check if files exist
	if _, err := os.Stat(lw.config.ConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", lw.config.ConfigPath)
	}

	if _, err := os.Stat(lw.config.DTOPath); os.IsNotExist(err) {
		return fmt.Errorf("DTO file not found: %s", lw.config.DTOPath)
	}

	// Generate swagger
	if err := generator.GenerateSwagger(
		lw.config.ConfigPath,
		lw.config.DTOPath,
		lw.config.OutputPath,
	); err != nil {
		return err
	}

	elapsed := time.Since(start)
	color.Green("✅ Generated in %s", elapsed)
	color.Cyan("📄 Output: %s", lw.config.OutputPath)
	color.Cyan("👀 Watching for changes...\n")

	return nil
}

func (lw *LiveWatcher) Stop() {
	lw.cancel()
}
