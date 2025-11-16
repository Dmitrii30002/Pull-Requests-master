package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`

	DB struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
		SSLMode  string
	}
	Logger struct {
		Level string `yaml:"level"`
		Path  string `yaml:"path"`
	} `yaml:"logger"`
}

func GetConfig() (*Config, error) {
	config := &Config{}
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	config.DB.Host = os.Getenv("DB_HOST")
	config.DB.Port = os.Getenv("DB_PORT")
	config.DB.Name = os.Getenv("DB_NAME")
	config.DB.User = os.Getenv("DB_USER")
	config.DB.Password = os.Getenv("DB_PASSWORD")
	config.DB.SSLMode = os.Getenv("DB_SSLMODE")

	return config, nil
}
