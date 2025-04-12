package builder

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/fanchann/docunyan/internals/models"
	"github.com/fanchann/docunyan/internals/utils"
)

// builds the complete OpenAPI specification
func BuildOpenAPISpec(doc models.DocunyanYAML, schemas map[string]interface{}) ([]byte, error) {
	// JSON Swagger
	swagger := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       doc.Info.Title,
			"version":     doc.Info.Version,
			"description": doc.Info.Description,
		},
		"paths": map[string]interface{}{},
		"components": map[string]interface{}{
			"schemas": schemas,
		},
	}

	// add servers if defined
	if len(doc.Servers) > 0 {
		servers := []map[string]interface{}{}
		for _, s := range doc.Servers {
			server := map[string]interface{}{"url": s.URL}
			if s.Description != "" {
				server["description"] = s.Description
			}
			servers = append(servers, server)
		}
		swagger["servers"] = servers
	}

	// security schemes to be populated if authorization is defined
	securitySchemes := make(map[string]interface{})
	globalSecurity := []map[string][]string{}

	if doc.Authorization != nil {
		components := swagger["components"].(map[string]interface{})

		// process each auth type
		for i, authType := range doc.Authorization.Type {
			authType = strings.ToLower(authType)
			var scheme string
			if i < len(doc.Authorization.Scheme) {
				scheme = strings.ToLower(doc.Authorization.Scheme[i])
			} else if len(doc.Authorization.Scheme) > 0 {
				scheme = strings.ToLower(doc.Authorization.Scheme[0])
			}

			securityKey := strings.Replace(authType+scheme, " ", "", -1)

			switch authType {
			case "http":
				securitySchemes[securityKey] = map[string]interface{}{
					"type":   "http",
					"scheme": scheme,
				}

				// add global security requirement
				globalSecurity = append(globalSecurity, map[string][]string{
					securityKey: {},
				})

			case "apikey":
				in := "header"
				if doc.Authorization.In != "" {
					in = strings.ToLower(doc.Authorization.In)
				}

				securitySchemes[securityKey] = map[string]interface{}{
					"type": "apiKey",
					"name": doc.Authorization.Name,
					"in":   in,
				}

				// add global security requirement
				globalSecurity = append(globalSecurity, map[string][]string{
					securityKey: {},
				})

			}
		}

		if len(securitySchemes) > 0 {
			components["securitySchemes"] = securitySchemes
			// add global security requirements - this will be overridden at endpoint level
			if len(globalSecurity) > 0 {
				swagger["security"] = globalSecurity
			}
		}
	}

	paths := buildPaths(doc.Paths, securitySchemes)
	swagger["paths"] = paths

	output, err := json.MarshalIndent(swagger, "", "  ")
	if err != nil {
		log.Printf("failed to marshal json: %v", err)
		return nil, err
	}

	return output, nil
}

// builds the paths section of the OpenAPI spec
func buildPaths(docPaths map[string]map[string]models.EndpointDetail, securitySchemes map[string]interface{}) map[string]interface{} {
	paths := map[string]interface{}{}

	for path, endpoints := range docPaths {
		pathItem := map[string]interface{}{}

		for method, endpoint := range endpoints {
			methodLower := strings.ToLower(method)

			// responses object
			responseObj := map[string]interface{}{}
			for code, resp := range endpoint.Responses {
				responseObj[code] = map[string]interface{}{
					"description": resp.Description,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/" + resp.Schema,
							},
						},
					},
				}
			}

			methodObj := map[string]interface{}{
				"summary":   endpoint.Summary,
				"responses": responseObj,
			}

			// add tags if present
			if len(endpoint.Tags) > 0 {
				methodObj["tags"] = endpoint.Tags
			}

			// Handle endpoint-specific authorization
			if endpoint.Authorization {
				// add security requirement to this endpoint
				if securitySchemes != nil && len(securitySchemes) > 0 {
					security := []map[string][]string{}

					// add all available security schemes to this endpoint
					for schemeName := range securitySchemes {
						security = append(security, map[string][]string{
							schemeName: {},
						})
					}

					if len(security) > 0 {
						methodObj["security"] = security
					}
				}
			} else {
				// if authorization is false (default), explicitly override global security
				// by providing an empty security requirement object
				methodObj["security"] = []map[string][]string{{}}
			}

			// Handle request body if specified
			if endpoint.RequestBody != "" {
				methodObj["requestBody"] = map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/" + endpoint.RequestBody,
							},
						},
					},
				}
			}

			// Handle parameters
			params := []map[string]interface{}{}

			// path parameters (extracted from path)
			pathParams := utils.ExtractPathParams(path)
			for _, param := range pathParams {
				params = append(params, map[string]interface{}{
					"name":     param,
					"in":       "path",
					"required": true,
					"schema":   map[string]interface{}{"type": "string"},
				})
			}

			// handle query parameters from the new query field
			if endpoint.Query != nil {
				for paramName, paramType := range endpoint.Query {
					paramObj := map[string]interface{}{
						"name":     paramName,
						"in":       "query",
						"required": false,
						"schema": map[string]interface{}{
							"type": utils.GoTypeToSwaggerType(paramType),
						},
					}
					params = append(params, paramObj)
				}
			}

			if endpoint.Parameter != nil {
				switch p := endpoint.Parameter.(type) {
				case string:
					// Simple type parameter
					params = append(params, map[string]interface{}{
						"name":     "body",
						"in":       "query", // Default to query
						"required": true,
						"schema":   map[string]interface{}{"type": utils.GoTypeToSwaggerType(p)},
					})
				case map[interface{}]interface{}:
					// Complex parameter object
					paramMap := utils.ConvertToStringMap(p)
					if name, ok := paramMap["name"].(string); ok {
						paramObj := map[string]interface{}{
							"name":     name,
							"in":       paramMap["in"].(string),
							"required": paramMap["required"].(bool),
						}
						if schema, ok := paramMap["schema"].(string); ok {
							if _, exists := models.StructSchemas[schema]; exists {
								paramObj["schema"] = map[string]interface{}{
									"$ref": "#/components/schemas/" + schema,
								}
							} else {
								paramObj["schema"] = map[string]interface{}{
									"type": utils.GoTypeToSwaggerType(schema),
								}
							}
						}
						params = append(params, paramObj)
					}
				}
			}

			// add explicitly defined parameters
			for _, param := range endpoint.Parameters {
				paramObj := map[string]interface{}{
					"name":     param.Name,
					"in":       param.In,
					"required": param.Required,
				}

				if param.Description != "" {
					paramObj["description"] = param.Description
				}

				schemaObj := map[string]interface{}{
					"type": param.Type,
				}

				// add format based on type
				switch param.Type {
				case "integer":
					schemaObj["format"] = "int64"
				case "number":
					schemaObj["format"] = "double"
				}

				paramObj["schema"] = schemaObj
				params = append(params, paramObj)
			}

			// add parameters if we have any
			if len(params) > 0 {
				methodObj["parameters"] = params
			}

			pathItem[methodLower] = methodObj
		}

		openAPIPath := utils.NormalizePathParams(path)
		paths[openAPIPath] = pathItem
	}

	return paths
}
