package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"reflect"
	"strings"

	"github.com/fanchann/docunyan/internals/utils"
)

type SchemaBuilder struct {
	Structs       map[string]*ast.StructType
	StructDocs    map[string]string
	StructSchemas map[string]map[string]interface{}
}

func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		Structs:       make(map[string]*ast.StructType),
		StructDocs:    make(map[string]string),
		StructSchemas: make(map[string]map[string]interface{}),
	}
}

func (s *SchemaBuilder) ParseGoStructs(contractFileName string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, contractFileName, nil, parser.ParseComments)
	if err != nil {
		log.Printf("failed to parse %s: %v", contractFileName, err)
		return err
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
				s.Structs[typeSpec.Name.Name] = structType
				if docComment != "" {
					s.StructDocs[typeSpec.Name.Name] = docComment
				}
			}
		}
	}
	return nil
}

func (s *SchemaBuilder) BuildSchemas() map[string]interface{} {
	processed := map[string]bool{}
	schemas := map[string]interface{}{}

	var processStruct func(name string)
	processStruct = func(name string) {
		if processed[name] {
			return
		}

		st := s.Structs[name]
		for _, field := range st.Fields.List {
			fieldType := utils.ExprToTypeString(field.Type)
			if strings.HasPrefix(fieldType, "[]") {
				elemType := strings.TrimPrefix(fieldType, "[]")
				if _, exists := s.Structs[elemType]; exists && !processed[elemType] {
					processStruct(elemType)
				}
			} else if _, exists := s.Structs[fieldType]; exists && !processed[fieldType] {
				processStruct(fieldType)
			}
		}

		s.StructSchemas[name] = s.parseStructSpec(name, st)
		schema := s.StructSchemas[name]

		if desc, ok := s.StructDocs[name]; ok && desc != "" {
			schema["description"] = strings.TrimSpace(desc)
		}

		schemas[name] = schema
		processed[name] = true
	}

	for name := range s.Structs {
		processStruct(name)
	}

	return schemas
}

func (s *SchemaBuilder) parseStructSpec(structName string, st *ast.StructType) map[string]interface{} {
	properties := map[string]interface{}{}
	required := []string{}

	for _, field := range st.Fields.List {
		var fieldName string
		if len(field.Names) > 0 {
			fieldName = field.Names[0].Name
		} else {
			// Embedded struct
			switch t := field.Type.(type) {
			case *ast.Ident:
				embeddedType := t.Name
				if embeddedSchema, ok := s.StructSchemas[embeddedType]; ok {
					if embProps, ok := embeddedSchema["properties"].(map[string]interface{}); ok {
						for k, v := range embProps {
							properties[k] = v
						}
					}
				}
			}
			continue
		}

		jsonTag := fieldName
		isRequired := true

		if tag := field.Tag; tag != nil {
			tagValue := reflect.StructTag(strings.Trim(tag.Value, "`"))
			if jsonKey := tagValue.Get("json"); jsonKey != "" {
				parts := strings.Split(jsonKey, ",")
				jsonTag = parts[0]
				if jsonTag == "-" {
					continue
				}
				for _, opt := range parts[1:] {
					if opt == "omitempty" {
						isRequired = false
						break
					}
				}
			}
			if validTag := tagValue.Get("validate"); validTag != "" {
				if strings.Contains(validTag, "required") {
					isRequired = true
				}
			}
		}

		if isRequired {
			required = append(required, jsonTag)
		}

		typeStr := utils.ExprToTypeString(field.Type)
		if strings.HasPrefix(typeStr, "[]") {
			elemType := strings.TrimPrefix(typeStr, "[]")
			if _, ok := s.StructSchemas[elemType]; ok {
				properties[jsonTag] = map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"$ref": "#/components/schemas/" + elemType},
				}
			} else {
				properties[jsonTag] = map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": utils.GoTypeToSwaggerType(elemType)},
				}
			}
		} else if _, ok := s.StructSchemas[typeStr]; ok {
			properties[jsonTag] = map[string]interface{}{
				"$ref": "#/components/schemas/" + typeStr,
			}
		} else {
			propSchema := map[string]interface{}{"type": utils.GoTypeToSwaggerType(typeStr)}
			if typeStr == "time.Time" {
				propSchema["format"] = "date-time"
			}
			properties[jsonTag] = propSchema
		}
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}
