# 🐱 Docunyan – Swagger Generator for Go Structs + YAML

**Docunyan** is a simple tool to generate Swagger/OpenAPI documentation from your Go structs and a YAML config. Ideal for teams building Go-based APIs with minimal overhead.

---

## ⚙️ Installation

```bash
go install github.com/fanchann/docunyan@latest
# or clone manually:
git clone https://github.com/fanchann/docunyan.git && cd docunyan && go install
```

---

## 🚀 Quick Start

### Option 1: Create New Project Folder

```bash
docunyan init --name my-api
cd my-api
docunyan watch
```

### Option 2: Initialize in Current Directory

```bash
mkdir my-api && cd my-api
docunyan init
docunyan watch
```

### What Gets Created:

```
my-api/
├── docunyan.yml          # API configuration
├── dto/
│   └── example.go        # Example DTOs
└── generated/
    └── generated.json    # Generated Swagger (after watch)
```

### Live Reload Features:

When you run `docunyan watch`:
- ✅ Auto-detect project structure
- ✅ Generate initial Swagger JSON
- ✅ Watch for file changes (docunyan.yml, dto/*.go)
- ✅ Auto-regenerate on changes
- ✅ Open Swagger UI in browser

### Add Your DTOs

Create any `.go` files in `dto/` directory:

```go
// dto/product.go
package dto

type Product struct {
    ID    int64  `json:"id"`
    Name  string `json:"name" validate:"required"`
    Price float64 `json:"price"`
}
```

**All `.go` files in `dto/` are automatically scanned!**

---

## 📋 Commands

### Main Commands

```bash
docunyan init                    # Initialize in current directory
docunyan init --name <folder>    # Create new project folder
docunyan watch                   # Start live reload (auto-detect)
docunyan --folder <path> watch   # Watch specific folder
docunyan validate [config.yml]   # Validate YAML config only
```

### Examples

```bash
# Create new project
docunyan init --name my-api
cd my-api && docunyan watch

# Initialize in current directory
mkdir my-api && cd my-api
docunyan init
docunyan watch

# Watch different folder
docunyan --folder ./api-v2 watch
```

### Legacy Commands (Still Supported)

```bash
docunyan --config <yml> --go-file <go|dir> [--output <json>]
docunyan --live <swagger.json>
```

---

## 📂 Recommended Structure

```
docs/
└── api/
    ├── product/
    │   ├── product.go
    │   ├── product.yml
    │   └── product.json
    └── auth/
        ├── auth.go
        ├── auth.yml
        └── auth.json
```

Use consistent naming: `example.go`, `example.yml`, `example.json`

---

## 📝 Config Example (`docunyan.yml`)

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
```

---

## 📌 Features

- 🚀 **Quick Init**: Bootstrap project with `docunyan init`
- 🔄 **Live Reload**: Auto-regenerate on file changes with `docunyan watch`
- 📁 **Directory Scanning**: Automatically scans all `.go` files in `dto/` folder
- 🔄 **Struct to Schema**: Convert Go structs into Swagger definitions
- 🔐 **Authorization**: Support for API keys & HTTP schemes
- 🔗 **Path & Query Parameters**: Full parameter support
- 📦 **Request Bodies**: Automatic schema generation
- 📊 **Live Preview**: Built-in Swagger UI with hot reload
- ✅ **Real-time Validation**: YAML config validation with error highlighting
- 🎯 **Auto-detect**: Smart project structure detection

---

## 🛠 Troubleshooting

- Export all structs (capitalize)
- Validate YAML formatting
- Ensure file paths are correct
- Match JSON tags to expected schema fields

---

## 💡 Best Practices

- Group endpoints with tags
- Separate request/response DTOs
- Always describe success & error responses
- Keep each feature in its own folder with matching file names
