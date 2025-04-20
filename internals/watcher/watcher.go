package watcher

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"

	"github.com/fanchann/docunyan/internals/models"
)

type ValidationError struct {
	Message string
	Line    int
	Field   string
}

type ConfigWatcher struct {
	filePath string
	watcher  *fsnotify.Watcher
	errors   []ValidationError
}

func NewConfigWatcher(filePath string) (*ConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %v", err)
	}

	return &ConfigWatcher{
		filePath: filePath,
		watcher:  watcher,
		errors:   make([]ValidationError, 0),
	}, nil
}

func (c *ConfigWatcher) Start() error {
	absPath, err := filepath.Abs(c.filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	c.filePath = absPath

	dir := filepath.Dir(c.filePath)
	err = c.watcher.Add(dir)
	if err != nil {
		return fmt.Errorf("failed to watch directory %s: %v", dir, err)
	}

	c.validateFile()

	fmt.Println()
	os.Stdout.WriteString("\033[H\033[2J")
	color.HiCyan("📝 Docunyan YAML Config Watcher")
	color.HiCyan("==============================")
	color.Cyan("Watching: %s", c.filePath)
	fmt.Println()

	for {
		select {
		case event, ok := <-c.watcher.Events:
			if !ok {
				return fmt.Errorf("watcher event channel closed")
			}

			if filepath.Clean(event.Name) == filepath.Clean(c.filePath) {
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					os.Stdout.WriteString("\033[H\033[2J")
					color.HiCyan("📝 Docunyan YAML Config Watcher")
					color.HiCyan("==============================")
					color.Cyan("Watching: %s", c.filePath)
					color.HiYellow("File changed at: %s", time.Now().Format("15:04:05"))
					fmt.Println()
					c.validateFile()
				}
			}

		case err, ok := <-c.watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher error channel closed")
			}
			color.Red("Watcher error: %v", err)
		}
	}
}

func (c *ConfigWatcher) Close() {
	if c.watcher != nil {
		c.watcher.Close()
	}
}

func (c *ConfigWatcher) validateFile() {
	start := time.Now()

	content, err := ioutil.ReadFile(c.filePath)
	if err != nil {
		color.Red("❌ Error reading file: %v", err)
		return
	}

	c.errors = make([]ValidationError, 0)

	var doc models.DocunyanYAML
	err = yaml.Unmarshal(content, &doc)
	if err != nil {
		errorMsg := err.Error()
		lineNum := 0

		if strings.Contains(errorMsg, "line") {
			parts := strings.Split(errorMsg, "line ")
			if len(parts) > 1 {
				fmt.Sscanf(parts[1], "%d", &lineNum)
			}
		}

		c.errors = append(c.errors, ValidationError{
			Message: fmt.Sprintf("YAML Parse Error: %v", err),
			Line:    lineNum,
		})
	} else {
		c.validateSchema(&doc, content)
	}

	if len(c.errors) > 0 {
		color.Red("❌ Validation failed with %d errors:", len(c.errors))

		lines := strings.Split(string(content), "\n")
		fmt.Println()
		color.Yellow("File content:")
		for i, line := range lines {
			lineNumber := i + 1

			hasError := false
			errorMessages := []string{}
			for _, err := range c.errors {
				if err.Line == lineNumber {
					hasError = true
					errorMessages = append(errorMessages, err.Message)
				}
			}

			if hasError {
				color.Red("%3d | %s", lineNumber, line)
				for _, msg := range errorMessages {
					color.Red("      └─ %s", msg)
				}
			} else {
				fmt.Printf("%3d | %s\n", lineNumber, line)
			}
		}

		fmt.Println()
		color.Red("Summary of errors:")
		for i, err := range c.errors {
			if err.Line > 0 {
				color.Red("%d. Line %d: %s", i+1, err.Line, err.Message)
			} else if err.Field != "" {
				color.Red("%d. Field '%s': %s", i+1, err.Field, err.Message)
			} else {
				color.Red("%d. %s", i+1, err.Message)
			}
		}
	} else {
		color.Green("✅ Configuration is valid!")

		fmt.Println()
		color.Yellow("Configuration Summary:")
		color.White("Title: %s", doc.Info.Title)
		color.White("Version: %s", doc.Info.Version)
		if doc.Info.Description != "" {
			color.White("Description: %s", doc.Info.Description)
		}

		color.White("Servers: %d", len(doc.Servers))
		color.White("Endpoints: %d", countEndpoints(doc.Paths))
		if doc.Authorization != nil {
			color.White("Auth Type: %s (%s)", doc.Authorization.Name, strings.Join(doc.Authorization.Type, ", "))
		}
	}

	elapsed := time.Since(start)
	fmt.Println()
	color.Cyan("Validation completed in %s", elapsed)
	fmt.Println()
}

func (c *ConfigWatcher) validateSchema(doc *models.DocunyanYAML, content []byte) {
	lines := strings.Split(string(content), "\n")
	lineMap := buildLineMap(lines)

	if doc.Info.Title == "" {
		c.addError("info.title", "Title is required", lineMap)
	}

	if doc.Info.Version == "" {
		c.addError("info.version", "Version is required", lineMap)
	}

	if len(doc.Paths) == 0 {
		c.addError("paths", "At least one path must be defined", lineMap)
	}

	if len(doc.Servers) == 0 {
		c.addError("servers", "At least one server must be defined", lineMap)
	} else {
		for i, server := range doc.Servers {
			if server.URL == "" {
				c.addError(fmt.Sprintf("servers[%d].url", i), "Server URL is required", lineMap)
			}
		}
	}

	for path, methods := range doc.Paths {
		if !strings.HasPrefix(path, "/") {
			c.addError(fmt.Sprintf("paths.%s", path), "Path must start with '/'", lineMap)
		}

		for method, detail := range methods {
			validMethods := []string{"get", "post", "put", "delete", "patch", "options", "head", "trace"}
			isValid := false
			for _, valid := range validMethods {
				if strings.ToLower(method) == valid {
					isValid = true
					break
				}
			}

			if !isValid {
				c.addError(fmt.Sprintf("paths.%s.%s", path, method),
					fmt.Sprintf("Invalid HTTP method: %s", method), lineMap)
			}

			if len(detail.Responses) == 0 {
				c.addError(fmt.Sprintf("paths.%s.%s.responses", path, method),
					"At least one response must be defined", lineMap)
			} else {
				for status, response := range detail.Responses {
					if !isValidStatusCode(status) {
						c.addError(fmt.Sprintf("paths.%s.%s.responses.%s", path, method, status),
							fmt.Sprintf("Invalid status code: %s", status), lineMap)
					}

					if response.Description == "" {
						c.addError(fmt.Sprintf("paths.%s.%s.responses.%s.description", path, method, status),
							"Description is required", lineMap)
					}

					if response.Schema == "" {
						c.addError(fmt.Sprintf("paths.%s.%s.responses.%s.schema", path, method, status),
							"Schema is required", lineMap)
					}
				}
			}

			if detail.Parameters != nil {
				for i, param := range detail.Parameters {
					if param.Name == "" {
						c.addError(fmt.Sprintf("paths.%s.%s.parameters[%d].name", path, method, i),
							"Parameter name is required", lineMap)
					}

					if param.In == "" {
						c.addError(fmt.Sprintf("paths.%s.%s.parameters[%d].in", path, method, i),
							"Parameter location (in) is required", lineMap)
					} else {
						validLocations := []string{"query", "path", "header", "cookie"}
						isValid := false
						for _, valid := range validLocations {
							if strings.ToLower(param.In) == valid {
								isValid = true
								break
							}
						}

						if !isValid {
							c.addError(fmt.Sprintf("paths.%s.%s.parameters[%d].in", path, method, i),
								fmt.Sprintf("Invalid 'in' value: %s", param.In), lineMap)
						}
					}

					if strings.ToLower(param.In) == "path" && !param.Required {
						c.addError(fmt.Sprintf("paths.%s.%s.parameters[%d].required", path, method, i),
							"Path parameters must be required", lineMap)
					}
				}
			}
		}
	}

	if doc.Authorization != nil {
		if doc.Authorization.Name == "" {
			c.addError("authorization.name", "Authorization name is required", lineMap)
		}

		if len(doc.Authorization.Type) == 0 {
			c.addError("authorization.type", "At least one authorization type must be specified", lineMap)
		}

		if len(doc.Authorization.Scheme) == 0 {
			c.addError("authorization.scheme", "At least one authorization scheme must be specified", lineMap)
		}
	}
}

func (c *ConfigWatcher) addError(field, message string, lineMap map[string]int) {
	lineNum := lineMap[field]
	c.errors = append(c.errors, ValidationError{
		Message: message,
		Line:    lineNum,
		Field:   field,
	})
}

func buildLineMap(lines []string) map[string]int {
	result := make(map[string]int)

	type StackItem struct {
		Path  string
		Level int
	}

	var stack []StackItem
	currentLevel := 0

	for i, line := range lines {
		lineNum := i + 1

		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		indentLevel := len(line) - len(strings.TrimLeft(line, " "))

		if len(stack) == 0 || indentLevel > currentLevel {
			parts := strings.SplitN(trimmedLine, ":", 2)
			if len(parts) < 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])

			if len(stack) == 0 {
				path := key
				stack = append(stack, StackItem{Path: path, Level: indentLevel})
				result[path] = lineNum
			} else {
				parentPath := stack[len(stack)-1].Path
				path := parentPath + "." + key
				stack = append(stack, StackItem{Path: path, Level: indentLevel})
				result[path] = lineNum
			}

			currentLevel = indentLevel
		} else if indentLevel < currentLevel {
			for len(stack) > 0 && stack[len(stack)-1].Level >= indentLevel {
				stack = stack[:len(stack)-1]
			}

			if len(stack) > 0 {
				currentLevel = stack[len(stack)-1].Level

				parts := strings.SplitN(trimmedLine, ":", 2)
				if len(parts) < 2 {
					continue
				}

				key := strings.TrimSpace(parts[0])
				parentPath := ""

				if len(stack) > 0 {
					pathParts := strings.Split(stack[len(stack)-1].Path, ".")
					parentPath = strings.Join(pathParts[:len(pathParts)-1], ".")
				}

				path := ""
				if parentPath == "" {
					path = key
				} else {
					path = parentPath + "." + key
				}

				stack = append(stack, StackItem{Path: path, Level: indentLevel})
				result[path] = lineNum
			}
		} else {
			parts := strings.SplitN(trimmedLine, ":", 2)
			if len(parts) < 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])

			if len(stack) > 0 {
				pathParts := strings.Split(stack[len(stack)-1].Path, ".")
				parentPath := strings.Join(pathParts[:len(pathParts)-1], ".")

				path := ""
				if parentPath == "" {
					path = key
				} else {
					path = parentPath + "." + key
				}

				stack[len(stack)-1] = StackItem{Path: path, Level: indentLevel}
				result[path] = lineNum
			}
		}
	}

	return result
}

func isValidStatusCode(status string) bool {
	if len(status) < 3 {
		return false
	}

	prefix := status[0]
	return prefix >= '1' && prefix <= '5'
}

func countEndpoints(paths map[string]map[string]models.EndpointDetail) int {
	count := 0
	for _, methods := range paths {
		count += len(methods)
	}
	return count
}
