package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fanchann/docunyan/internals/generator"
	"github.com/fanchann/docunyan/internals/live"
)

func Execute() {
	configPath := flag.String("config", "", "Path to docunyan.yml")
	goFilePath := flag.String("go-file", "", "Path to Go file")
	outputPath := flag.String("output", "", "Output Swagger file (optional)")
	livePreview := flag.String("live", "", "Swagger file for live preview (optional)")

	flag.Parse()

	// live preview only
	if *livePreview != "" && *configPath == "" && *goFilePath == "" {
		live.SwaggerLive(*livePreview)
		return
	}

	// generate only (with optional live)
	if *configPath == "" || *goFilePath == "" {
		fmt.Println("Usage:")
		fmt.Println("  Generate: docunyan --config path/to/docunyan.yml --go-file path/to/response.go [--output path/to/swagger.yaml] [--live path/to/swagger.yaml]")
		fmt.Println("  Preview:  docunyan --live path/to/swagger.yaml")
		os.Exit(1)
	}

	err := generator.GenerateSwagger(*configPath, *goFilePath, *outputPath)
	if err != nil {
		log.Fatalf("Failed to generate Swagger: %v\n", err)
	}

	// Optional live preview after generate
	if *livePreview != "" {
		live.SwaggerLive(*livePreview)
	}
}
