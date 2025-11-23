package watcher

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/fanchann/docunyan/internals/constants"
	"github.com/fanchann/docunyan/internals/errors"
	"github.com/fanchann/docunyan/internals/generator"
	"github.com/fanchann/docunyan/internals/live"
	"github.com/fanchann/docunyan/internals/utils"
)

// LiveWatcherConfig defines configuration for the live watcher
type LiveWatcherConfig struct {
	RootDir    string // Root directory to watch (default: current directory)
	ConfigFile string // Config file name (default: docunyan.yml)
	DTODir     string // DTO directory (default: dto)
	OutputFile string // Output file (default: generated/generated.json)
}

// LiveWatcher monitors file changes and triggers regeneration
type LiveWatcher struct {
	config       LiveWatcherConfig
	watcher      *fsnotify.Watcher
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	previewCtx   context.Context
	previewCancel context.CancelFunc
	debounceCh   chan string
}

// FileChange represents a file change event
type FileChange struct {
	Path string
	Op   fsnotify.Op
}

// progressStatus tracks regeneration progress
type progressStatus struct {
	stage     string
	timestamp time.Time
	complete  bool
}

func DefaultLiveWatcherConfig() LiveWatcherConfig {
	return LiveWatcherConfig{
		RootDir:    ".",
		ConfigFile: "docunyan.yml",
		DTODir:     "dto",
		OutputFile: "generated/generated.json",
	}
}

// NewLiveWatcher creates a new LiveWatcher instance with proper resource management
func NewLiveWatcher(config LiveWatcherConfig) (*LiveWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.NewWatcherError("", "failed to create file system watcher", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	previewCtx, previewCancel := context.WithCancel(context.Background())

	// Resolve absolute paths
	absRoot, err := filepath.Abs(config.RootDir)
	if err != nil {
		cancel()
		previewCancel()
		watcher.Close()
		return nil, errors.NewWatcherError(config.RootDir, "failed to resolve root directory", err)
	}

	config.RootDir = absRoot

	return &LiveWatcher{
		config:       config,
		watcher:      watcher,
		ctx:          ctx,
		cancel:       cancel,
		previewCtx:   previewCtx,
		previewCancel: previewCancel,
		debounceCh:   make(chan string, 10), // Buffered channel
	}, nil
}

// cleanup performs graceful shutdown of all resources
func (lw *LiveWatcher) cleanup() {
	// Cancel all contexts first
	lw.cancel()
	lw.previewCancel()

	// Close debounce channel
	close(lw.debounceCh)

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		lw.wg.Wait()
		close(done)
	}()

	// Wait with timeout
	select {
	case <-done:
		// All goroutines finished cleanly
	case <-time.After(5 * time.Second):
		// Force cleanup if timeout
		fmt.Fprintf(os.Stderr, "Warning: Force shutdown - some goroutines did not exit cleanly\n")
	}

	// Close the watcher
	if lw.watcher != nil {
		lw.watcher.Close()
	}
}

// Start begins the file watching process with proper goroutine management
func (lw *LiveWatcher) Start() error {
	defer lw.cleanup()

	// Build full paths
	configPath := filepath.Join(lw.config.RootDir, lw.config.ConfigFile)
	dtoPath := filepath.Join(lw.config.RootDir, lw.config.DTODir)
	outputPath := filepath.Join(lw.config.RootDir, lw.config.OutputFile)

	// Validate paths exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return errors.NewConfigError(configPath, "configuration file not found", nil)
	}

	if _, err := os.Stat(dtoPath); os.IsNotExist(err) {
		return errors.NewConfigError(dtoPath, "DTO directory not found", nil)
	}

	// Watch root directory (for config file)
	if err := lw.watcher.Add(lw.config.RootDir); err != nil {
		return errors.NewWatcherError(lw.config.RootDir, "failed to watch root directory", err)
	}

	// Watch DTO directory
	if err := lw.watcher.Add(dtoPath); err != nil {
		return errors.NewWatcherError(dtoPath, "failed to watch DTO directory", err)
	}

	// Initial generation
	utils.PrintHeader("DOCUNYAN LIVE WATCHER v2.0")
	fmt.Println()
	utils.PrintProjectInfo(lw.config.RootDir, lw.config.ConfigFile, lw.config.DTODir, lw.config.OutputFile)

	if err := lw.regenerate(configPath, dtoPath, outputPath); err != nil {
		utils.PrintError("Initial generation failed: %v", err)
		return err
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start debounced file processor
	lw.wg.Add(1)
	go lw.startDebouncedProcessor(configPath, dtoPath, outputPath)

	// Start live preview in managed goroutine
	lw.wg.Add(1)
	go lw.startPreviewServer(outputPath)

	// Main event loop
	for {
		select {
		case <-lw.ctx.Done():
			utils.PrintShutdown()
			return nil

		case <-sigChan:
			utils.PrintShutdown()
			lw.cancel()
			return nil

		case event, ok := <-lw.watcher.Events:
			if !ok {
				return errors.NewWatcherError("", "watcher event channel closed", nil)
			}

			if lw.shouldRegenerate(event, configPath, dtoPath) {
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					// Send to debounce channel instead of creating new goroutine
					select {
					case lw.debounceCh <- filepath.Base(event.Name):
					default:
						// Channel full, skip this event
					}
				}
			}

		case err, ok := <-lw.watcher.Errors:
			if !ok {
				return errors.NewWatcherError("", "watcher error channel closed", nil)
			}
			utils.PrintWarning("Watcher error: %v", err)
		}
	}
}

// startDebouncedProcessor handles file change events with debouncing
func (lw *LiveWatcher) startDebouncedProcessor(configPath, dtoPath, outputPath string) {
	defer lw.wg.Done()

	var lastEvent time.Time
	var timer *time.Timer

	for {
		select {
		case <-lw.ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return

		case filename, ok := <-lw.debounceCh:
			if !ok {
				return // Channel closed
			}

			now := time.Now()
			timeSinceLastEvent := now.Sub(lastEvent)

			// Reset existing timer
			if timer != nil {
				timer.Stop()
			}

			if timeSinceLastEvent >= constants.FileWatcherDebounce {
				// Enough time has passed, regenerate immediately
				lw.handleFileChange(filename, configPath, dtoPath, outputPath)
				lastEvent = now
			} else {
				// Wait for remaining debounce time
				remainingTime := constants.FileWatcherDebounce - timeSinceLastEvent
				timer = time.AfterFunc(remainingTime, func() {
					select {
					case <-lw.ctx.Done():
						return
					default:
						lw.handleFileChange(filename, configPath, dtoPath, outputPath)
						lastEvent = time.Now()
					}
				})
			}
		}
	}
}

// startPreviewServer starts the live preview server with proper lifecycle management
func (lw *LiveWatcher) startPreviewServer(outputPath string) {
	defer lw.wg.Done()

	// Wait a bit for initial generation
	select {
	case <-time.After(500 * time.Millisecond):
	case <-lw.previewCtx.Done():
		return
	}

	// Check if output file exists
	if _, err := os.Stat(outputPath); err != nil {
		return // File doesn't exist, skip preview
	}

	// Start preview server
	if err := live.SwaggerLiveWithContext(lw.previewCtx, outputPath); err != nil {
		select {
		case <-lw.previewCtx.Done():
			// Expected shutdown
		default:
			utils.PrintWarning("Live preview server error: %v", err)
		}
	}
}

// handleFileChange processes a single file change event
func (lw *LiveWatcher) handleFileChange(filename, configPath, dtoPath, outputPath string) {
	select {
	case <-lw.ctx.Done():
		return
	default:
		utils.PrintFileChange(filename)

		if err := lw.regenerate(configPath, dtoPath, outputPath); err != nil {
			utils.PrintError("Regeneration failed: %v", err)
		}
	}
}

// shouldRegenerate determines if a file change should trigger regeneration
func (lw *LiveWatcher) shouldRegenerate(event fsnotify.Event, configPath, dtoPath string) bool {
	eventFile := filepath.Clean(event.Name)
	configFile := filepath.Clean(configPath)

	// Check if it's the config file
	if eventFile == configFile {
		return true
	}

	// Check if it's a .go file in the DTO directory (not test files)
	// Use more efficient file extension checking
	if filepath.Ext(eventFile) == ".go" && !strings.HasSuffix(eventFile, "_test.go") {
		eventDir := filepath.Clean(filepath.Dir(eventFile))
		dtoDirClean := filepath.Clean(dtoPath)
		return eventDir == dtoDirClean
	}

	return false
}

// regenerate triggers swagger generation with real progress tracking
func (lw *LiveWatcher) regenerate(configPath, dtoPath, outputPath string) error {
	start := time.Now()

	// Create progress tracker
	progress := make(chan progressStatus, 4)
	defer close(progress)

	// Start progress display goroutine
	lw.wg.Add(1)
	go func() {
		defer lw.wg.Done()
		lw.displayProgress(progress)
	}()

	// Send initial progress
	progress <- progressStatus{stage: "Scanning Go files...", timestamp: time.Now()}

	// Stage 1: Validate files exist
	if _, err := os.Stat(configPath); err != nil {
		return errors.NewConfigError(configPath, "configuration file not found", err)
	}

	progress <- progressStatus{stage: "Parsing structs...", timestamp: time.Now()}

	// Stage 2: Generate swagger (this is the actual work)
	if err := generator.GenerateSwagger(configPath, dtoPath, outputPath); err != nil {
		return errors.NewGenerateError(outputPath, "failed to generate swagger documentation", err)
	}

	// Send completion
	progress <- progressStatus{stage: "Complete", timestamp: time.Now(), complete: true}

	elapsed := time.Since(start)
	utils.PrintGenerated(elapsed)
	utils.PrintInfo("Output: %s", outputPath)
	utils.PrintWatching()

	return nil
}

// displayProgress shows regeneration progress without artificial delays
func (lw *LiveWatcher) displayProgress(progress <-chan progressStatus) {
	s := utils.NewGeneratingSpinner()
	s.Suffix = "Initializing..."
	s.Start()
	defer s.Stop()

	for status := range progress {
		if status.complete {
			break
		}
		s.Suffix = status.stage
	}
}

func (lw *LiveWatcher) Stop() {
	lw.cancel()
}

// AutoDetectConfig tries to auto-detect the project structure
func AutoDetectConfig(rootDir string) (LiveWatcherConfig, error) {
	config := LiveWatcherConfig{
		RootDir:    rootDir,
		ConfigFile: "docunyan.yml",
		DTODir:     "dto",
		OutputFile: "generated/generated.json",
	}

	// Check if config file exists
	configPath := filepath.Join(rootDir, config.ConfigFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, fmt.Errorf("config file not found: %s (run 'docunyan init' first)", configPath)
	}

	// Check if DTO directory exists
	dtoPath := filepath.Join(rootDir, config.DTODir)
	if _, err := os.Stat(dtoPath); os.IsNotExist(err) {
		return config, fmt.Errorf("DTO directory not found: %s", dtoPath)
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Join(rootDir, filepath.Dir(config.OutputFile))
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, constants.DefaultDirPermission); err != nil {
			return config, fmt.Errorf("failed to create output directory: %w", err)
		}
		log.Printf("Created output directory: %s\n", outputDir)
	}

	return config, nil
}
