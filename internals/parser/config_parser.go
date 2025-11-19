package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/fanchann/docunyan/internals/builder"
	"github.com/fanchann/docunyan/internals/models"
)

func DocunyanConfigParser(docunyanConf string, contractFileName string) ([]byte, error) {
	var doc models.DocunyanYAML
	yamlFile, err := os.ReadFile(docunyanConf)
	if err != nil {
		return nil, fmt.Errorf("failed to read docunyan.yml: %w", err)
	}
	if err := yaml.Unmarshal(yamlFile, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	schemaBuilder := NewSchemaBuilder()

	// parse Go structs
	if err := schemaBuilder.ParseGoStructs(contractFileName); err != nil {
		return nil, fmt.Errorf("failed to parse Go structs: %w", err)
	}

	// build schemas from structs
	schemas := schemaBuilder.BuildSchemas()

	output, err := builder.BuildOpenAPISpec(doc, schemas)
	if err != nil {
		return nil, fmt.Errorf("failed to build OpenAPI spec: %w", err)
	}

	return output, nil
}
