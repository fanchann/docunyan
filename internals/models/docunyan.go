package models

// DocunyanYAML represents the main configuration structure for API documentation
type DocunyanYAML struct {
	Info struct {
		Title       string `yaml:"title"`
		Version     string `yaml:"version"`
		Description string `yaml:"description,omitempty"`
	} `yaml:"info"`
	Servers []struct {
		URL         string `yaml:"url"`
		Description string `yaml:"description,omitempty"`
	} `yaml:"servers"`
	Paths         map[string]map[string]EndpointDetail `yaml:"paths"`
	Authorization *Authorization                       `yaml:"authorization,omitempty"`
}

// Authorization defines security schemes for the API
type Authorization struct {
	Name   string   `yaml:"name"`
	Type   []string `yaml:"type"`
	Scheme []string `yaml:"scheme"`
	In     string   `yaml:"in,omitempty"`
}

// Parameter defines endpoint parameters
type Parameter struct {
	Name        string `yaml:"name,omitempty"`
	In          string `yaml:"in,omitempty"`
	Required    bool   `yaml:"required,omitempty"`
	Type        string `yaml:"type,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// Response defines endpoint responses
type Response struct {
	Description string `yaml:"description"`
	Schema      string `yaml:"schema"`
}

// EndpointDetail contains all information about an API endpoint
type EndpointDetail struct {
	Query         map[string]string   `yaml:"query,omitempty"`
	Summary       string              `yaml:"summary,omitempty"`
	Tags          []string            `yaml:"tags,omitempty"`
	RequestBody   string              `yaml:"requestBody,omitempty"`
	Parameter     interface{}         `yaml:"parameter,omitempty"`
	Parameters    []Parameter         `yaml:"parameters,omitempty"`
	Responses     map[string]Response `yaml:"responses,omitempty"`
	Authorization bool                `yaml:"authorization,omitempty"`
}
