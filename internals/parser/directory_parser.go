package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ParseGoDirectory parses all .go files in a directory and extracts structs
func ParseGoDirectory(dirPath string) (*SchemaBuilder, error) {
	schemaBuilder := NewSchemaBuilder()

	// Check if path is a directory
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		// If it's a single file, parse it directly
		if err := schemaBuilder.ParseGoStructs(dirPath); err != nil {
			return nil, err
		}
		return schemaBuilder, nil
	}

	// Read all files in directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	filesFound := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .go files, skip test files
		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		if err := parseGoFile(schemaBuilder, filePath); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", entry.Name(), err)
		}
		filesFound++
	}

	if filesFound == 0 {
		return nil, fmt.Errorf("no .go files found in directory: %s", dirPath)
	}

	return schemaBuilder, nil
}

// parseGoFile parses a single Go file and adds structs to the schema builder
func parseGoFile(sb *SchemaBuilder, filePath string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		docComment := ""
		if genDecl.Doc != nil {
			docComment = genDecl.Doc.Text()
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				// Check if struct already exists (avoid duplicates)
				if _, exists := sb.Structs[typeSpec.Name.Name]; !exists {
					sb.Structs[typeSpec.Name.Name] = structType
					if docComment != "" {
						sb.StructDocs[typeSpec.Name.Name] = docComment
					}
				}
			}
		}
	}

	return nil
}
