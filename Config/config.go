package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		Port     int    `yaml:"port"`
		SSLMode  string `yaml:"sslmode"`
		TimeZone string `yaml:"timezone"`
	} `yaml:"database"`
	Server struct {
		Port           string `yaml:"port"`
		ReadTimeout    int    `yaml:"read_timeout"`
		WriteTimeout   int    `yaml:"write_timeout"`
		MaxHeaderBytes int    `yaml:"max_header_bytes"`
	} `yaml:"server"`
}

func LoadConfig(filename string) (*Config, error) {
	config := &Config{}
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml file: %w", err)
	}
	return config, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		c.Database.Host,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.Port,
		c.Database.SSLMode,
		c.Database.TimeZone)
}
