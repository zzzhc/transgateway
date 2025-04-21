package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Providers struct {
		Mtranserver struct {
			Enable    bool     `yaml:"enable"`
			Endpoints []string `yaml:"endpoints"`
		} `yaml:"mtranserver"`
		Google struct {
			Enable bool   `yaml:"enable"`
			Proxy  string `yaml:"proxy"`
		} `yaml:"google"`
	} `yaml:"providers"`
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
