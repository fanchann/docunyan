package generator

import (
	"fmt"
	"os"
	"time"

	"github.com/fanchann/docunyan/internals/parser"
)

func GenerateSwagger(configPath, goFilePath, outputPath string) error {
	outputTempl, err := parser.DocunyanConfigParser(configPath, goFilePath)
	if err != nil {
		return fmt.Errorf("failed while parsing configuration: %w", err)
	}

	if outputPath == "" {
		outputPath = fmt.Sprintf("docunyan_gen_%s.json", time.Now().Format("20060102_150405"))
	}

	if err := os.WriteFile(outputPath, outputTempl, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("✅ Swagger generated successfully at %s\n", outputPath)
	return nil
}
