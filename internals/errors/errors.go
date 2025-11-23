package errors

import (
	"fmt"
	"strings"
)

// DocunyanError represents domain-specific errors with context
type DocunyanError struct {
	Op      string // Operation being performed
	Path    string // File or resource path
	Message string // Human-readable message
	Cause   error  // Underlying error
}

// Error implements the error interface
func (e *DocunyanError) Error() string {
	var parts []string

	if e.Op != "" {
		parts = append(parts, e.Op)
	}

	if e.Path != "" {
		parts = append(parts, e.Path)
	}

	if e.Message != "" {
		parts = append(parts, e.Message)
	}

	errMsg := strings.Join(parts, ": ")
	if e.Cause != nil {
		errMsg += fmt.Sprintf(" (%v)", e.Cause)
	}

	return errMsg
}

// Unwrap returns the underlying error
func (e *DocunyanError) Unwrap() error {
	return e.Cause
}

// NewDocunyanError creates a new DocunyanError
func NewDocunyanError(op, path, message string, cause error) *DocunyanError {
	return &DocunyanError{
		Op:      op,
		Path:    path,
		Message: message,
		Cause:   cause,
	}
}

// Predefined error constructors
func NewConfigError(path, message string, cause error) error {
	return NewDocunyanError("config", path, message, cause)
}

func NewParseError(path, message string, cause error) error {
	return NewDocunyanError("parser", path, message, cause)
}

func NewGenerateError(path, message string, cause error) error {
	return NewDocunyanError("generator", path, message, cause)
}

func NewValidationError(path, message string, cause error) error {
	return NewDocunyanError("validation", path, message, cause)
}

func NewFileSystemError(path, message string, cause error) error {
	return NewDocunyanError("filesystem", path, message, cause)
}

func NewWatcherError(path, message string, cause error) error {
	return NewDocunyanError("watcher", path, message, cause)
}

func NewServerError(path, message string, cause error) error {
	return NewDocunyanError("server", path, message, cause)
}

// ValidationErrors aggregates multiple validation errors
type ValidationErrors []error

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}

	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Error())
	}

	if len(ve) == 1 {
		return "validation error: " + messages[0]
	}

	return fmt.Sprintf("validation errors (%d): %s", len(ve), strings.Join(messages, "; "))
}

// Add adds a new validation error
func (ve *ValidationErrors) Add(err error) {
	if err != nil {
		*ve = append(*ve, err)
	}
}

// HasErrors returns true if there are any validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Any returns true if any error matches the predicate
func (ve ValidationErrors) Any(predicate func(error) bool) bool {
	for _, err := range ve {
		if predicate(err) {
			return true
		}
	}
	return false
}

// PathTraversalError represents a path traversal security error
type PathTraversalError struct {
	Path     string
	Attempted string
}

// Error implements the error interface
func (e *PathTraversalError) Error() string {
	return fmt.Sprintf("path traversal detected: attempted to access '%s' from base '%s'", e.Attempted, e.Path)
}

// NewPathTraversalError creates a new path traversal error
func NewPathTraversalError(base, attempted string) error {
	return &PathTraversalError{
		Path:      base,
		Attempted: attempted,
	}
}

// ConfigurationError represents configuration-specific errors
type ConfigurationError struct {
	Field   string
	Value   string
	Message string
}

// Error implements the error interface
func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("configuration error: field '%s' with value '%s' - %s", e.Field, e.Value, e.Message)
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(field, value, message string) error {
	return &ConfigurationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}