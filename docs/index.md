---
layout: default
title: Docunyan Documentation
---

## 🐱 Docunyan – Swagger Generator from Go Structs + YAML
**Docunyan** is a documentation generator for Swagger/OpenAPI, designed specifically for Go applications. It simplifies the process of creating API documentation by automatically generating Swagger/OpenAPI specifications from your Go structs and configuration files.

---

## ⚙️ Installation

```bash
# Install via go install
go install github.com/fanchann/docunyan@latest

# Or clone the repository manually
git clone https://github.com/fanchann/docunyan.git
cd docunyan
go install
```

---

## 🚀 Basic Usage

The primary command format for Docunyan is:

```bash
docunyan --config <path/to/docunyan.yml> --go-file <path/to/dto.go> [--output <path/to/swagger.yaml>] [--live <path/to/swagger.yaml>]
```

### 🔑 Core Parameters

- `--config`: Path to the Docunyan YAML config file **(required)**
- `--go-file`: Path to the Go file containing request/response structs **(required)**
- `--output`: Destination to save the generated Swagger/OpenAPI spec *(optional)*
- `--live`: Start a live Swagger UI preview of the documentation *(optional)*

### 👀 Preview Mode

To preview your API documentation with Swagger UI **without writing a file**:

```bash
docunyan --live <path/to/swagger.yaml>
```

This will start a local web server with Swagger UI for real-time visualization.

---

## 🗂️ Folder Structure & Naming Conventions

Docunyan expects the YAML configuration and the Go file to be in the **same folder**, making file references simpler and more consistent.

### ✅ Recommended Structure

```
└── docs
	├── api/
	│   ├── product/
	│   │   ├── product.go      # Go DTO file
	│   │   ├── product.yml     # Docunyan config
	│   │   └── product.json    # Generated Swagger output
	│   ├── order/
	│   │   ├── order.go
	│   │   ├── order.yml
	│   │   └── order.json
	│   └── auth/
	│       ├── auth.go
	│       ├── auth.yml
	│       └── auth.json
```

### 📛 Naming Convention

Keep naming consistent for ease of use:

```
├── example1.go     # Go DTO
├── example1.yml    # Docunyan config
└── example1.json   # Swagger output
```

Docunyan automatically links the config and Go file from the same folder.

### 🧪 Example Command

```bash
# Generate documentation from a specific folder
docunyan --config ./api/product/product.yml --go-file ./api/product/product.go --output ./api/product/product.json

# Or from inside the product directory:
cd ./api/product
docunyan --config product.yml --go-file product.go --output product.json
```

---

## 📝 Configuration File Structure

Here’s an example of a `docunyan.yml` file:

```yaml
info:
  title: Your API Title
  version: 1.0.0
  description: Description of your API

servers:
  - url: http://localhost:8080/api
    description: Development server

authorization:
  name: X-API-KEY
  type: [apiKey]
  in: header

paths:
  /your/endpoint:
    get:
      authorization: true
      summary: Endpoint description
      tags: [Category]
      query:
        paramName: type
      responses:
        200:
          description: Success response
          schema: YourResponseType
        400:
          description: Bad request
          schema: ErrorResponse
```

### 🔍 Key Sections

1. **Info**: API metadata
2. **Servers**: Environment endpoints
3. **Authorization**: API auth settings
4. **Paths**: API routes, parameters, and responses

---

## 🔄 Mapping Go Structs to API Schemas

Docunyan extracts type definitions from your Go file to build Swagger schemas.

```go
type ProductResponse struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    Price       float64 `json:"price"`
    Description string  `json:"description"`
}
```

Referenced in YAML like this:

```yaml
paths:
  /products:
    get:
      responses:
        200:
          description: Product list
          schema: ProductResponse
```

---

## ✨ Advanced Features

### 🔗 Path Parameters

```yaml
paths:
  /products/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          type: string
          description: Product ID
      responses:
        200:
          description: Product details
          schema: ProductDetailResponse
```

### 🔎 Query Parameters

```yaml
paths:
  /products:
    get:
      query:
        page: int
        pageSize: int
        search: string
      responses:
        200:
          description: Paginated products
          schema: ProductListResponse
```

### 📦 Request Bodies

```yaml
paths:
  /products:
    post:
      requestBody: CreateProductRequest
      responses:
        201:
          description: Product created
          schema: ProductResponse
```

### 🔐 Authorization

```yaml
authorization:
  name: X-API-KEY
  type: [apiKey]
  in: header

paths:
  /public/endpoint:
    get:
      authorization: false  # Public

  /protected/endpoint:
    get:
      authorization: true   # Requires API key
```

---

## 📚 Examples

### 1. Simple API

```yaml
info:
  title: Simple API
  version: 1.0.0

servers:
  - url: http://localhost:8080/api

paths:
  /hello:
    get:
      summary: Say hello
      responses:
        200:
          description: Hello response
          schema: HelloResponse
```

### 2. Authenticated API

```yaml
info:
  title: Protected API
  version: 1.0.0

servers:
  - url: http://localhost:8080/api

authorization:
  name: X-API-KEY
  type: [apiKey]
  in: header

paths:
  /secure/resource:
    get:
      authorization: true
      summary: Get secure resource
      responses:
        200:
          description: Success
          schema: SecureResource
        401:
          description: Unauthorized
          schema: ErrorResponse
```

### 3. Full CRUD Example

**Folder:**
```
project/
└── docs
	└── api/
		└── product/
        	├── product.go
        	├── product.yml
        	└── product.json
```

**product.yml**
```yaml
info:
  title: Product API
  version: 1.0.0

servers:
  - url: http://localhost:8080/api

authorization:
  name: X-API-KEY
  type: [apiKey]
  in: header

paths:
  /products:
    get:
      summary: List products
      query:
        page: int
        pageSize: int
      responses:
        200:
          description: Product list
          schema: ProductListResponse

    post:
      summary: Create product
      requestBody: CreateProductRequest
      responses:
        201:
          description: Product created
          schema: ProductResponse
```

**product.go**
```go
package product

type ProductListResponse struct {
    Success bool              `json:"success"`
    Data    []ProductResponse `json:"data"`
}

type ProductResponse struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type CreateProductRequest struct {
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}
```

---

## 🛠️ Troubleshooting

### Common Issues

1. **Missing schema**: Ensure your Go struct is exported (capitalized) and properly tagged.
2. **Path not visible**: Check YAML syntax for indentation and correctness.
3. **Validation errors**: Make sure JSON tags and data types in Go match Swagger expectations.
4. **File not found**: Ensure config and Go files are in the same directory.

---

## 💡 Best Practices

- **Group by Tags**: Categorize endpoints with consistent tags
- **Cover All Responses**: Always document both success and error cases
- **Write Descriptions**: Add meaningful descriptions to endpoints and parameters
- **Dedicated DTOs**: Avoid reusing the same struct for request/response
- **Clear Naming**: Use descriptive names for paths and schemas
- **Organize by Folder**: Keep config, DTO, and output files in the same folder
- **Be Consistent**: Follow a consistent naming strategy

---