package utils

import (
	"go/ast"
	"strings"
)

func GenerateQueryParameters(queryParamsDef interface{}) []map[string]interface{} {
	if queryParamsDef == nil {
		return nil
	}

	params := []map[string]interface{}{}

	switch qp := queryParamsDef.(type) {
	case map[interface{}]interface{}:
		paramMap := ConvertToStringMap(qp)
		for name, details := range paramMap {
			param := CreateParameterObject(name, details, "query")
			if param != nil {
				params = append(params, param)
			}
		}
	case map[string]interface{}:
		for name, details := range qp {
			param := CreateParameterObject(name, details, "query")
			if param != nil {
				params = append(params, param)
			}
		}
	case []interface{}:
		for _, item := range qp {
			if paramMap, ok := item.(map[interface{}]interface{}); ok {
				convertedMap := ConvertToStringMap(paramMap)
				if name, exists := convertedMap["name"].(string); exists {
					param := CreateParameterObject(name, convertedMap, "query")
					if param != nil {
						params = append(params, param)
					}
				}
			}
		}
	}

	return params
}

//  creates a parameter object for OpenAPI spec
func CreateParameterObject(name string, details interface{}, paramIn string) map[string]interface{} {
	paramObj := map[string]interface{}{
		"name":     name,
		"in":       paramIn,
		"required": false, // Default to false
	}

	switch d := details.(type) {
	case map[string]interface{}:
		// Extract common properties
		if required, ok := d["required"].(bool); ok {
			paramObj["required"] = required
		}
		if description, ok := d["description"].(string); ok {
			paramObj["description"] = description
		}

		// Handle schema based on type
		schemaObj := map[string]interface{}{}

		if typeVal, ok := d["type"].(string); ok {
			schemaObj["type"] = typeVal

			// Add format based on type
			switch typeVal {
			case "integer":
				if format, ok := d["format"].(string); ok {
					schemaObj["format"] = format
				} else {
					schemaObj["format"] = "int64"
				}
			case "number":
				if format, ok := d["format"].(string); ok {
					schemaObj["format"] = format
				} else {
					schemaObj["format"] = "double"
				}
			case "string":
				if format, ok := d["format"].(string); ok {
					schemaObj["format"] = format
				}
			case "object":
				// Handle nested object properties
				if properties, ok := d["properties"].(map[string]interface{}); ok {
					schemaObj["properties"] = properties
				}
			}
		} else {
			// Default to string if type is not specified
			schemaObj["type"] = "string"
		}

		// Handle enum values
		if enum, ok := d["enum"].([]interface{}); ok && len(enum) > 0 {
			schemaObj["enum"] = enum
		}

		// Handle default value
		if defaultVal, ok := d["default"]; ok {
			schemaObj["default"] = defaultVal
		}

		paramObj["schema"] = schemaObj

	case string:
		// Simple type parameter
		paramObj["schema"] = map[string]interface{}{
			"type": GoTypeToSwaggerType(d),
		}
	}

	return paramObj
}

//  converts map[interface{}]interface{} to map[string]interface{}
func ConvertToStringMap(m map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		switch key := k.(type) {
		case string:
			switch val := v.(type) {
			case map[interface{}]interface{}:
				result[key] = ConvertToStringMap(val)
			case []interface{}:
				result[key] = ConvertToStringSlice(val)
			default:
				result[key] = v
			}
		}
	}
	return result
}

//  converts []interface{} to proper string or map slice
func ConvertToStringSlice(a []interface{}) []interface{} {
	result := make([]interface{}, len(a))
	for i, v := range a {
		switch val := v.(type) {
		case map[interface{}]interface{}:
			result[i] = ConvertToStringMap(val)
		case []interface{}:
			result[i] = ConvertToStringSlice(val)
		default:
			result[i] = v
		}
	}
	return result
}

func ExprToTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return "[]" + ExprToTypeString(t.Elt)
	case *ast.StarExpr:
		return ExprToTypeString(t.X)
	case *ast.SelectorExpr:
		x := ExprToTypeString(t.X)
		return x + "." + t.Sel.Name
	case *ast.MapType:
		keyType := ExprToTypeString(t.Key)
		valueType := ExprToTypeString(t.Value)
		return "map[" + keyType + "]" + valueType
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return "unknown"
	}
}

// converts a Go type to a Swagger/OpenAPI type
func GoTypeToSwaggerType(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "int", "int64", "int32", "int16", "int8", "uint", "uint64", "uint32", "uint16", "uint8":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "time.Time":
		return "string"
	case "interface{}":
		return "object"
	default:
		if strings.HasPrefix(goType, "map[") {
			return "object"
		}
		return "object"
	}
}
