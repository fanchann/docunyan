package parser

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/fanchann/docunyan/internals/builder"
	"github.com/fanchann/docunyan/internals/models"
)

func DocunyanConfigParser(docunyanConf string, contractFileName string) ([]byte, error) {
	var doc models.DocunyanYAML
	yamlFile, err := os.ReadFile(docunyanConf)
	if err != nil {
		log.Fatalf("Failed to read docunyan.yml: %v", err)
		return nil, err
	}
	if err := yaml.Unmarshal(yamlFile, &doc); err != nil {
		log.Fatalf("Failed to unmarshal yaml: %v", err)
		return nil, err
	}

	schemaBuilder := NewSchemaBuilder()

	// parse Go structs
	if err := schemaBuilder.ParseGoStructs(contractFileName); err != nil {
		log.Fatalf("Failed to parse Go structs: %v", err)
		return nil, err
	}

	// build schemas from structs
	schemas := schemaBuilder.BuildSchemas()

	output, err := builder.BuildOpenAPISpec(doc, schemas)
	if err != nil {
		log.Fatalf("Failed to build OpenAPI spec: %v", err)
		return nil, err
	}

	return output, nil
}
