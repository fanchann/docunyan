package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fanchann/docunyan/internals/generator"
	"github.com/fanchann/docunyan/internals/live"
	"github.com/fanchann/docunyan/internals/watcher"
)

var docunyanLogo = `
 /\_/\
( o o )
 =_Y_=
  '-'
`

func Execute() {
	configPath := flag.String("config", "", "Path to docunyan.yml")
	goFilePath := flag.String("go-file", "", "Path to Go file")
	outputPath := flag.String("output", "", "Output Swagger file (optional)")
	livePreview := flag.String("live", "", "Swagger file for live preview (optional)")
	watchFile := flag.String("watcher", "", "Path to YAML file for live validation")

	flag.Parse()

	switch {
	case *livePreview != "" && *configPath == "" && *goFilePath == "" && *watchFile == "":
		live.SwaggerLive(*livePreview)
		return

	case *watchFile != "" && *configPath == "" && *goFilePath == "" && *livePreview == "":
		w, err := watcher.NewConfigWatcher(*watchFile)
		if err != nil {
			log.Fatalf("Failed to Watch File: %v\n", err)
		}
		w.Start()
		return

	case *configPath != "" && *goFilePath != "":
		err := generator.GenerateSwagger(*configPath, *goFilePath, *outputPath)
		if err != nil {
			log.Fatalf("Failed to generate Swagger: %v\n", err)
		}

		if *livePreview != "" {
			live.SwaggerLive(*livePreview)
		}
		return
	}

	fmt.Println(docunyanLogo)
	fmt.Println("Usage:")
	fmt.Println("  Generate Swagger:         docunyan --config path/to/docunyan.yml --go-file path/to/response.go [--output path/to/swagger.json] [--live path/to/swagger.yaml]")
	fmt.Println("  Live Preview Only:        docunyan --live path/to/swagger.yaml")
	fmt.Println("  Realtime YAML Validation: docunyan --watcher path/to/docunyan.yml")
	os.Exit(1)
}
