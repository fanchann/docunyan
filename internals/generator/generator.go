package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/briandowns/spinner"

	"github.com/fanchann/docunyan/internals/constants"
	"github.com/fanchann/docunyan/internals/errors"
	"github.com/fanchann/docunyan/internals/parser"
)

// GenerateSwagger generates OpenAPI specification from config and Go files
func GenerateSwagger(configPath, goFilePath, outputPath string) error {
	// Input validation
	if configPath == "" {
		return errors.NewValidationError("", "config path cannot be empty", nil)
	}
	if goFilePath == "" {
		return errors.NewValidationError("", "Go file path cannot be empty", nil)
	}

	// Validate file paths exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return errors.NewConfigError(configPath, "configuration file not found", err)
	}

	if _, err := os.Stat(goFilePath); os.IsNotExist(err) {
		return errors.NewParseError(goFilePath, "Go file/directory not found", err)
	}

	outputTempl, err := parser.DocunyanConfigParser(configPath, goFilePath)
	if err != nil {
		return errors.NewParseError(configPath, "failed to parse configuration", err)
	}

	if outputPath == "" {
		outputPath = fmt.Sprintf(constants.DefaultOutputPattern, time.Now().Format("20060102_150405"))
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, constants.DefaultDirPermission); err != nil {
		return errors.NewFileSystemError(outputDir, "failed to create output directory", err)
	}

	if err := os.WriteFile(outputPath, outputTempl, constants.DefaultFilePermission); err != nil {
		return errors.NewFileSystemError(outputPath, "failed to write output file", err)
	}

	return nil
}

// GenerateSwaggerWithMessage generates swagger with user feedback through spinner
func GenerateSwaggerWithMessage(configPath, goFilePath, outputPath string) error {
	// Input validation with enhanced error messages
	if configPath == "" {
		fmt.Fprintf(os.Stderr, "Error: Config path cannot be empty\n")
		return errors.NewValidationError("", "config path cannot be empty", nil)
	}
	if goFilePath == "" {
		fmt.Fprintf(os.Stderr, "Error: Go file path cannot be empty\n")
		return errors.NewValidationError("", "Go file path cannot be empty", nil)
	}

	// Start spinner with stages
	s := newSpinner()
	s.Suffix = " Scanning Go files..."
	s.Start()

	// Stage 1: Scanning (no artificial delay)
	s.Suffix = " Parsing structs..."

	// Stage 2: Parsing (no artificial delay)
	s.Suffix = " Building OpenAPI schemas..."

	// Stage 3: Building (no artificial delay)
	s.Suffix = " Generating Swagger documentation..."

	err := GenerateSwagger(configPath, goFilePath, outputPath)

	s.Stop()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Generation failed: %v\n", err)
		return err
	}

	fmt.Printf("Success: Swagger generated successfully at %s\n", outputPath)
	return nil
}

// newSpinner creates a new spinner instance with consistent configuration
func newSpinner() *spinner.Spinner {
	s := spinner.New(spinner.CharSets[constants.SpinnerCharSetIndex], constants.SpinnerInterval)
	s.Color(constants.SpinnerColor)
	return s
}
