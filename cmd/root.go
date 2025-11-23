package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/fanchann/docunyan/internals/errors"
	"github.com/fanchann/docunyan/internals/initializer"
	"github.com/fanchann/docunyan/internals/watcher"
)

var docunyanLogo = `
 /\_/\
( o o )
 =_Y_=
  '-'
`

func Execute() {
	// Check for subcommands first
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			handleInit()
			return
		case "watch":
			folder := "."
			if len(os.Args) > 2 && os.Args[2] == "--folder" && len(os.Args) > 3 {
				folder = os.Args[3]
			}
			handleWatch(folder)
			return
		case "validate":
			configFile := "docunyan.yml"
			if len(os.Args) > 2 {
				configFile = os.Args[2]
			}
			handleValidate(configFile)
			return
		case "--folder":
			// Handle: docunyan --folder <path> watch
			if len(os.Args) > 3 && os.Args[3] == "watch" {
				handleWatch(os.Args[2])
				return
			}
		case "--help", "-h":
			showHelp()
			return
		case "--version", "-v":
			showVersion()
			return
		}
	}

	// Show help by default
	showHelp()
}

func handleInit() {
	// Parse init flags
	initFlags := flag.NewFlagSet("init", flag.ExitOnError)
	projectName := initFlags.String("name", "", "Project name (creates folder)")
	if err := initFlags.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", errors.NewValidationError("init", "failed to parse init flags", err))
		os.Exit(1)
	}

	// Validate project name
	if *projectName != "" {
		if len(*projectName) > 100 {
			fmt.Fprintf(os.Stderr, "Error: %v\n", errors.NewValidationError("init", "project name too long (max 100 characters)", nil))
			os.Exit(1)
		}
	}

	var config initializer.InitConfig

	if *projectName != "" {
		// Create project in new folder
		config = initializer.InitConfig{
			ConfigPath: *projectName + "/docunyan.yml",
			DTOPath:    *projectName + "/dto",
			OutputPath: *projectName + "/generated/generated.json",
		}
	} else {
		// Create in current directory
		config = initializer.DefaultInitConfig()
	}

	if err := initializer.Initialize(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *projectName != "" {
		fmt.Printf("\nNext: cd %s && docunyan watch\n", *projectName)
	} else {
		fmt.Println("\nNext: Run 'docunyan watch' to start live reload")
	}
}

func handleWatch(folder string) {
	// Validate folder path
	if folder == "" {
		fmt.Fprintf(os.Stderr, "Error: %v\n", errors.NewValidationError("watch", "folder path cannot be empty", nil))
		os.Exit(1)
	}

	// Auto-detect configuration
	config, err := watcher.AutoDetectConfig(folder)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create and start live watcher
	lw, err := watcher.NewLiveWatcher(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := lw.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleValidate(configFile string) {
	// Validate config file path
	if configFile == "" {
		fmt.Fprintf(os.Stderr, "Error: %v\n", errors.NewValidationError("validate", "config file path cannot be empty", nil))
		os.Exit(1)
	}

	w, err := watcher.NewConfigWatcher(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := w.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println(docunyanLogo)
	fmt.Println("Docunyan - Swagger/OpenAPI Generator for Go")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  docunyan <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init [--name <folder>]     Initialize new project")
	fmt.Println("  watch                      Start live reload (auto-detect)")
	fmt.Println("  --folder <path> watch      Start live reload in specific folder")
	fmt.Println("  validate [config.yml]      Validate YAML configuration")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --help, -h                 Show this help message")
	fmt.Println("  --version, -v              Show version information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  docunyan init               # Create in current directory")
	fmt.Println("  docunyan init --name my-api # Create my-api/ folder")
	fmt.Println("  docunyan watch              # Watch current directory")
	fmt.Println("  docunyan --folder ./api watch # Watch ./api directory")
	fmt.Println("  docunyan validate           # Validate current directory")
	fmt.Println()
	fmt.Println("Getting Started:")
	fmt.Println("  1. Create new project:  docunyan init --name my-api")
	fmt.Println("  2. Add your DTOs to:   my-api/dto/")
	fmt.Println("  3. Start live preview:  cd my-api && docunyan watch")
	fmt.Println()
	os.Exit(0)
}

func showVersion() {
	fmt.Println("Docunyan v2.0")
	fmt.Println("Swagger/OpenAPI Generator for Go")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  • Live reload with WebSocket preview")
	fmt.Println("  • Auto-detect project structure")
	fmt.Println("  • Generate from Go structs + YAML")
	fmt.Println("  • Zero-configuration setup")
	fmt.Println()
	os.Exit(0)
}
