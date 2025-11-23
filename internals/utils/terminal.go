package utils

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

// Status symbols
const (
	SymbolSuccess = "[✓]"
	SymbolError   = "[✗]"
	SymbolInfo    = "[i]"
	SymbolWarning = "[!]"
	SymbolArrow   = "[→]"
	SymbolDot     = "[•]"
)

// PrintSuccess prints a success message
func PrintSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	color.Green("%s %s", SymbolSuccess, msg)
}

// PrintError prints an error message
func PrintError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	color.Red("%s %s", SymbolError, msg)
}

// PrintInfo prints an info message
func PrintInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	color.Cyan("%s %s", SymbolInfo, msg)
}

// PrintWarning prints a warning message
func PrintWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	color.Yellow("%s %s", SymbolWarning, msg)
}

// PrintHeader prints a header with box drawing
func PrintHeader(title string) {
	color.HiCyan("╔════════════════════════════════════════╗")
	color.HiCyan("║  %-36s  ║", title)
	color.HiCyan("╚════════════════════════════════════════╝")
}

// PrintSeparator prints a separator line
func PrintSeparator() {
	fmt.Println("──────────────────────────────────────────────────")
}

// NewSpinner creates a new spinner with custom message
func NewSpinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Color("cyan")
	return s
}

// NewGeneratingSpinner creates a spinner for generation process
func NewGeneratingSpinner() *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Generating Swagger documentation..."
	s.Color("yellow")
	return s
}

// NewScanningSpinner creates a spinner for scanning process
func NewScanningSpinner() *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Scanning Go files..."
	s.Color("cyan")
	return s
}

// PrintBanner prints the docunyan banner
func PrintBanner() {
	color.HiCyan("╔════════════════════════════════════════╗")
	color.HiCyan("║         DOCUNYAN v2.0                  ║")
	color.HiCyan("║   Swagger Generator for Go APIs        ║")
	color.HiCyan("╚════════════════════════════════════════╝")
}

// PrintFileChange prints file change notification
func PrintFileChange(filename string) {
	color.HiYellow("\n%s File changed: %s", SymbolDot, filename)
}

// PrintGenerating prints generating message
func PrintGenerating() {
	color.HiYellow("%s Regenerating...", SymbolArrow)
}

// PrintGenerated prints generation complete message with duration
func PrintGenerated(duration time.Duration) {
	color.Green("%s Generated in %s", SymbolSuccess, duration)
}

// PrintWatching prints watching message
func PrintWatching() {
	color.Cyan("%s Watching for changes...\n", SymbolInfo)
}

// PrintShutdown prints shutdown message
func PrintShutdown() {
	color.Yellow("\n%s Shutting down gracefully...", SymbolInfo)
}

// PrintProjectInfo prints project information
func PrintProjectInfo(root, config, dto, output string) {
	color.Cyan("%s Root:   %s", SymbolInfo, root)
	color.Cyan("%s Config: %s", SymbolInfo, config)
	color.Cyan("%s DTOs:   %s", SymbolInfo, dto)
	color.Cyan("%s Output: %s", SymbolInfo, output)
	fmt.Println()
}
