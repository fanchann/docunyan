package models

var (
	StructSchemas = map[string]map[string]interface{}{}
)

type Parameter struct {
	Name        string `yaml:"name,omitempty"`
	In          string `yaml:"in,omitempty"`
	Required    bool   `yaml:"required,omitempty"`
	Type        string `yaml:"type,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type Response struct {
	Description string `yaml:"description"`
	Schema      string `yaml:"schema"`
}

type EndpointDetail struct {
	Summary     string              `yaml:"summary,omitempty"`
	Tags        []string            `yaml:"tags,omitempty"`
	RequestBody string              `yaml:"requestBody,omitempty"`
	Parameter   interface{}         `yaml:"parameter,omitempty"` // Can be string or complex object
	Parameters  []Parameter         `yaml:"parameters,omitempty"`
	Responses   map[string]Response `yaml:"responses,omitempty"`
}

type DocunyanYAML struct {
	Info struct {
		Title       string `yaml:"title"`
		Version     string `yaml:"version"`
		Description string `yaml:"description,omitempty"`
	} `yaml:"info"`
	Servers []struct {
		URL         string `yaml:"url"`
		Description string `yaml:"description,omitempty"`
	} `yaml:"servers,omitempty"`
	Paths map[string]map[string]EndpointDetail `yaml:"paths"`
}
