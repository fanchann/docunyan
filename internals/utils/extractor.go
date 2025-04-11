package utils

import (
	"log"
	"net/url"
	"strconv"
	"strings"
)

// converts REST-style path params to OpenAPI style
// e.g., "/users/:id" -> "/users/{id}"
func NormalizePathParams(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := strings.TrimPrefix(part, ":")
			parts[i] = "{" + paramName + "}"
		}
	}
	return strings.Join(parts, "/")
}

// extracts path parameters from a path string
// e.g., "/users/:id" -> ["id"]
func ExtractPathParams(path string) []string {
	parts := strings.Split(path, "/")
	params := []string{}

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			params = append(params, strings.TrimPrefix(part, ":"))
		}
	}

	return params
}

// parses URL query parameters into appropriate types
func ParseQueryParams(queryString string) map[string]interface{} {
	if queryString == "" {
		return nil
	}

	parsedValues := make(map[string]interface{})
	values, err := url.ParseQuery(queryString)
	if err != nil {
		log.Printf("Error parsing query string: %v", err)
		return nil
	}

	for key, vals := range values {
		if len(vals) == 0 {
			continue
		}

		// Check if there are nested parameters
		if strings.Contains(vals[0], "=") {
			// Handle nested parameters
			nestedParams := make(map[string]interface{})
			nestedVals, err := url.ParseQuery(vals[0])
			if err == nil {
				for nestedKey, nestedVal := range nestedVals {
					if len(nestedVal) > 0 {
						nestedParams[nestedKey] = InferType(nestedVal[0])
					}
				}
				parsedValues[key] = nestedParams
			} else {
				// if parsing failed, just use the raw value
				parsedValues[key] = vals[0]
			}
		} else if len(vals) > 1 {
			// handle array parameters
			typedArray := make([]interface{}, len(vals))
			for i, val := range vals {
				typedArray[i] = InferType(val)
			}
			parsedValues[key] = typedArray
		} else {
			// single value parameter
			parsedValues[key] = InferType(vals[0])
		}
	}

	return parsedValues
}

//  infers the Go type from a string value
func InferType(val string) interface{} {
	// Try to parse as bool
	if val == "true" {
		return true
	}
	if val == "false" {
		return false
	}

	// Try to parse as int
	if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {
		return intVal
	}

	// Try to parse as float
	if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
		return floatVal
	}

	// Default to string
	return val
}
