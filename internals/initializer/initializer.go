package initializer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"

	"github.com/fanchann/docunyan/internals/constants"
)

type InitConfig struct {
	ConfigPath string
	DTOPath    string // Can be a file or directory
	OutputPath string
}

func DefaultInitConfig() InitConfig {
	return InitConfig{
		ConfigPath: "docunyan.yml",
		DTOPath:    "dto", // Now points to directory
		OutputPath: "generated/generated.json",
	}
}

// Initialize creates the initial project structure with template files
func Initialize(config InitConfig) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Creating project directories..."
	s.Color("cyan")
	s.Start()

	time.Sleep(400 * time.Millisecond)

	// Create root directory if it's part of the path (for --name flag)
	rootDir := filepath.Dir(config.ConfigPath)
	if rootDir != "" && rootDir != "." {
		if err := createDirectory(rootDir); err != nil {
			s.Stop()
			return fmt.Errorf("failed to create project directory: %w", err)
		}
	}

	// Create DTO directory (config.DTOPath is now a directory)
	if err := createDirectory(config.DTOPath); err != nil {
		s.Stop()
		return fmt.Errorf("failed to create dto directory: %w", err)
	}

	// Create generated directory
	if err := createDirectory(filepath.Dir(config.OutputPath)); err != nil {
		s.Stop()
		return fmt.Errorf("failed to create generated directory: %w", err)
	}

	s.Suffix = " Generating configuration files..."
	time.Sleep(600 * time.Millisecond)

	// Create docunyan.yml
	if err := createFileIfNotExists(config.ConfigPath, DocunyanYAMLTemplate); err != nil {
		s.Stop()
		return fmt.Errorf("failed to create docunyan.yml: %w", err)
	}

	s.Suffix = " Creating example DTOs..."
	time.Sleep(500 * time.Millisecond)

	// Create example DTO file in the dto directory
	exampleFilePath := filepath.Join(config.DTOPath, "example.go")
	if err := createFileIfNotExists(exampleFilePath, ExampleDTOTemplate); err != nil {
		s.Stop()
		return fmt.Errorf("failed to create example DTO: %w", err)
	}

	time.Sleep(300 * time.Millisecond)
	s.Stop()

	// Print success messages
	if rootDir != "" && rootDir != "." {
		fmt.Printf("[✓] Created project directory: %s\n", rootDir)
	}
	fmt.Printf("[✓] Created %s\n", config.ConfigPath)
	fmt.Printf("[✓] Created %s\n", exampleFilePath)
	fmt.Printf("[✓] Created %s directory\n", filepath.Dir(config.OutputPath))

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Project initialized successfully!")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("\nNext steps:")
	fmt.Printf("  1. Review and customize %s\n", config.ConfigPath)
	fmt.Printf("  2. Add your DTOs to %s/ directory\n", config.DTOPath)
	fmt.Printf("  3. Run: docunyan watch\n")
	fmt.Println("\n[i] Tip: All .go files in dto/ will be automatically scanned!")

	return nil
}

func createDirectory(path string) error {
	if path == "" || path == "." {
		return nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, constants.DefaultDirPermission); err != nil {
			return err
		}
	}
	return nil
}

func createFileIfNotExists(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file already exists: %s", path)
	}

	return os.WriteFile(path, []byte(content), constants.DefaultFilePermission)
}
