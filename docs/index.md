# 🐱 Docunyan – Swagger Generator from Go Structs + YAML

**Docunyan** is a lightweight CLI tool that generates a complete **Swagger (OpenAPI)** schema from your Go structs (`dto.go`) and a YAML configuration file (`docunyan.yml`).  

🎯 Built for developers who want clean documentation **without boilerplate**, **no `$ref`**, and **no annotations** in code.

---

## 📦 Installation

```bash
go install github.com/fanchann/docunyan@latest
```

Or build it manually:

```bash
git clone https://github.com/fanchann/docunyan.git
cd docunyan
go build -o docunyan .
```

---

## 🗂️ Minimal Project Structure

```
.
├── docunyan.yml  # OpenAPI configuration file
└── dto.go        # Go structs with JSON tags
```

---

## ⚙️ How to Use

```bash
docunyan --config docunyan.yml --go-file dto.go --output swagger.json
```

### CLI Options

| Flag          | Description                                    | Default        |
|---------------|------------------------------------------------|----------------|
| `--config`    | Path to the YAML configuration file           | `docunyan.yml` |
| `--go-file`   | Path to the Go file containing DTOs           | *required*     |
| `--output`    | Output file (Swagger JSON)               | `""`       |
| `--live`    	| Swagger live preview               | *required*       |



---

## ✍️ Sample `docunyan.yml`

```yaml
info:
  title: Example API
  version: 2.0.0

servers:
  - url: localhost:8080/api

authorization:
  name: X-API-KEY
  type: [apiKey]
  in: header

paths:
  /products:
    get:
      authorization: true
      summary: Get paginated product list
      tags: [Product]
      query:
        page: int
        pageSize: int
        search: string
      responses:
        200:
          description: List of products with pagination
          schema: ProductListResponse
        401:
          description: Unauthorized
          schema: ErrorResponse
```

---

## 🧱 Sample DTOs (`dto.go`)

```go
type Product struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	Categories  []Category `json:"categories"`
	Available   bool       `json:"available"`
}

type ProductListResponse struct {
	Success    bool       `json:"success"`
	Pagination Pagination `json:"pagination"`
	Data       []Product  `json:"data"`
}

type ErrorResponse struct{
	ErrorCode string `json:"error_code"`
	Message string `json:"message"`
}
```

Supports:
- Nested structs
- Arrays, maps, primitives
- Inline schema generation (no `$ref`!)

---

## 📤 Example Output

Run:

```bash
docunyan -c docunyan.yml -g dto.go -o swagger.json
```

Docunyan will generate a full `swagger.json` file with:

- All schemas inlined
- Query parameters and responses injected from YAML
- Authorization handled automatically

---

## 🌟 Why Docunyan?

✅ No annotations required  
✅ Clean, fully inlined Swagger output  
✅ Go-first design philosophy  
✅ Simple config + powerful output  
---

## 🤝 Contributing

Want to contribute? We’d love your help!  
Feel free to open issues or submit PRs at:  
👉 [github.com/fanchann/docunyan](https://github.com/fanchann/docunyan)

---