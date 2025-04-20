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

```bash
docunyan --config <docunyan.yml> --go-file <dto.go> [--output <swagger.yaml>] [--live <swagger.yaml>]
```

### Required Flags
- `--config`: YAML config file path
- `--go-file`: Go DTO file path

### Optional Flags
- `--output`: Save generated Swagger file
- `--live`: Start local Swagger UI preview

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

- 🔄 **Struct to Schema**: Convert Go structs into Swagger definitions
- 🔐 **Authorization**: Support for API keys
- 🔗 **Path & Query Parameters**
- 📦 **Request Bodies** handling
- 📊 **Live Preview** via Swagger UI

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
