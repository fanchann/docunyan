package parser

import (
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/fanchann/docunyan/internals/builder"
	"github.com/fanchann/docunyan/internals/errors"
	"github.com/fanchann/docunyan/internals/models"
)

// DocunyanConfigParser parses YAML configuration and Go files to generate OpenAPI specification
func DocunyanConfigParser(docunyanConf string, contractFileName string) ([]byte, error) {
	// Input validation
	if docunyanConf == "" {
		return nil, errors.NewValidationError("", "docunyan configuration path cannot be empty", nil)
	}
	if contractFileName == "" {
		return nil, errors.NewValidationError("", "contract file/directory path cannot be empty", nil)
	}

	// Validate file extension for config
	if !strings.HasSuffix(docunyanConf, ".yml") && !strings.HasSuffix(docunyanConf, ".yaml") {
		return nil, errors.NewConfigError(docunyanConf, "docunyan configuration must be a .yml or .yaml file", nil)
	}

	// Check if config file exists
	if _, err := os.Stat(docunyanConf); os.IsNotExist(err) {
		return nil, errors.NewConfigError(docunyanConf, "configuration file not found", err)
	}

	// Check if contract file/directory exists
	if _, err := os.Stat(contractFileName); os.IsNotExist(err) {
		return nil, errors.NewParseError(contractFileName, "contract file/directory not found", err)
	}

	// Parse YAML configuration
	var doc models.DocunyanYAML
	yamlFile, err := os.ReadFile(docunyanConf)
	if err != nil {
		return nil, errors.NewConfigError(docunyanConf, "failed to read configuration", err)
	}

	if err := yaml.Unmarshal(yamlFile, &doc); err != nil {
		return nil, errors.NewConfigError(docunyanConf, "failed to parse YAML configuration", err)
	}

	// Validate required fields in YAML
	if doc.Info.Title == "" {
		return nil, errors.NewConfigurationError("title", "", "API title is required in configuration")
	}
	if doc.Info.Version == "" {
		return nil, errors.NewConfigurationError("version", "", "API version is required in configuration")
	}

	// Parse Go structs from file or directory
	schemaBuilder, err := ParseGoDirectory(contractFileName)
	if err != nil {
		return nil, errors.NewParseError(contractFileName, "failed to parse Go structs", err)
	}

	// Build schemas from structs
	schemas := schemaBuilder.BuildSchemas()

	output, err := builder.BuildOpenAPISpec(doc, schemas)
	if err != nil {
		return nil, errors.NewGenerateError("", "failed to build OpenAPI specification", err)
	}

	return output, nil
}
