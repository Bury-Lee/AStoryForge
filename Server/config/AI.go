package config

// AI defines one model persona configuration.
type AI struct {
	Enable      bool    `yaml:"enable" json:"enable"`           // 是否启用,不启用的话当analysis调用MCP的话自动移交人工
	Name        string  `yaml:"name" json:"name"`               // Persona name shown in logs and prompts.
	Prompt      string  `yaml:"prompt"`                         // Persona prompt template.
	Model       string  `yaml:"model" json:"model"`             // Model name or local model identifier.
	Temperature float32 `yaml:"temperature" json:"temperature"` // Sampling temperature.
	MaxTokens   int     `yaml:"max_tokens" json:"max_tokens"`   // Maximum response token count.
	Host        string  `yaml:"host" json:"host"`               // API endpoint for the model service.
	ApiKey      string  `yaml:"apiKey" json:"-"`                // API key for the model service.
	APIType     string  `yaml:"apiType" json:"apiType"`         // API provider type, such as openai.
}
