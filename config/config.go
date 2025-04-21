package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ProviderConfig struct {
	Enable       bool     `yaml:"enable" default:"true"`
	LLM          bool     `yaml:"llm,omitempty"`
	BaseUrl      string   `yaml:"baseUrl,omitempty"`
	ApiKey       string   `yaml:"apiKey,omitempty"`
	Model        string   `yaml:"model,omitempty"`
	SystemPrompt string   `yaml:"system_prompt,omitempty"`
	UserPrompt   string   `yaml:"user_prompt,omitempty"`
	Endpoints    []string `yaml:"endpoints,omitempty"`
	Proxy        string   `yaml:"proxy,omitempty"`
}

type Config struct {
	Host      string                    `yaml:"host"`
	Port      int                       `yaml:"port"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
