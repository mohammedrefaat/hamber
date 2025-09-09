package config

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	//"golang.org/x/oauth2/apple"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Database  DatabaseConfig  `yaml:"database"`
	Server    ServerConfig    `yaml:"server"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	OAuth     OAuthConfigs    `yaml:"oauth"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	Port     int    `yaml:"port"`
	SSLMode  string `yaml:"sslmode"`
	TimeZone string `yaml:"timezone"`
}

type ServerConfig struct {
	Port           string `yaml:"port"`
	ReadTimeout    string `yaml:"read_timeout"`  // Changed to string for YAML parsing
	WriteTimeout   string `yaml:"write_timeout"` // Changed to string for YAML parsing
	MaxHeaderBytes int    `yaml:"max_header_bytes"`
}

type RateLimitConfig struct {
	Requests int    `yaml:"requests"`
	Window   string `yaml:"window"` // Changed to string for YAML parsing
}

type OAuthConfigs struct {
	Google   OAuthProviderConfig `yaml:"google"`
	Facebook OAuthProviderConfig `yaml:"facebook"`
	//Apple    OAuthProviderConfig `yaml:"apple"`
}

type OAuthProviderConfig struct {
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url"`
	Scopes       []string `yaml:"scopes"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", filename, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	cfg.applyDefaults()

	return &cfg, nil
}

// GetDSN returns the Postgres connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		c.Database.Host,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.Port,
		c.Database.SSLMode,
		c.Database.TimeZone,
	)
}

// GetReadTimeout returns the read timeout as time.Duration
func (c *Config) GetReadTimeout() time.Duration {
	if c.Server.ReadTimeout == "" {
		return 5 * time.Second // default
	}
	duration, err := time.ParseDuration(c.Server.ReadTimeout)
	if err != nil {
		return 5 * time.Second // fallback to default
	}
	return duration
}

// GetWriteTimeout returns the write timeout as time.Duration
func (c *Config) GetWriteTimeout() time.Duration {
	if c.Server.WriteTimeout == "" {
		return 10 * time.Second // default
	}
	duration, err := time.ParseDuration(c.Server.WriteTimeout)
	if err != nil {
		return 10 * time.Second // fallback to default
	}
	return duration
}

// GetRateWindow returns the rate limit window as time.Duration
func (c *Config) GetRateWindow() time.Duration {
	if c.RateLimit.Window == "" {
		return time.Minute // default
	}
	duration, err := time.ParseDuration(c.RateLimit.Window)
	if err != nil {
		return time.Minute // fallback to default
	}
	return duration
}

// Validate basic sanity of config
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server.port must not be empty")
	}
	if c.Database.Host == "" || c.Database.User == "" || c.Database.DBName == "" {
		return fmt.Errorf("database host, user and dbname must not be empty")
	}

	// Validate duration strings
	if c.Server.ReadTimeout != "" {
		if _, err := time.ParseDuration(c.Server.ReadTimeout); err != nil {
			return fmt.Errorf("invalid read_timeout format: %w", err)
		}
	}
	if c.Server.WriteTimeout != "" {
		if _, err := time.ParseDuration(c.Server.WriteTimeout); err != nil {
			return fmt.Errorf("invalid write_timeout format: %w", err)
		}
	}
	if c.RateLimit.Window != "" {
		if _, err := time.ParseDuration(c.RateLimit.Window); err != nil {
			return fmt.Errorf("invalid rate limit window format: %w", err)
		}
	}

	return nil
}

// applyDefaults sets safe defaults where config values are missing
func (c *Config) applyDefaults() {
	if c.Server.ReadTimeout == "" {
		c.Server.ReadTimeout = "5s"
	}
	if c.Server.WriteTimeout == "" {
		c.Server.WriteTimeout = "10s"
	}
	if c.Server.MaxHeaderBytes == 0 {
		c.Server.MaxHeaderBytes = 1 << 20 // 1 MB
	}
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
	if c.Database.TimeZone == "" {
		c.Database.TimeZone = "UTC"
	}
	if c.RateLimit.Requests == 0 {
		c.RateLimit.Requests = 100
	}
	if c.RateLimit.Window == "" {
		c.RateLimit.Window = "1m"
	}
}

// GetOAuthConfig returns initialized OAuth configurations
func (c *Config) GetOAuthConfig() *OAuthConfig {
	return &OAuthConfig{
		Google: &oauth2.Config{
			ClientID:     c.OAuth.Google.ClientID,
			ClientSecret: c.OAuth.Google.ClientSecret,
			RedirectURL:  c.OAuth.Google.RedirectURL,
			Scopes:       c.OAuth.Google.Scopes,
			Endpoint:     google.Endpoint,
		},
		Facebook: &oauth2.Config{
			ClientID:     c.OAuth.Facebook.ClientID,
			ClientSecret: c.OAuth.Facebook.ClientSecret,
			RedirectURL:  c.OAuth.Facebook.RedirectURL,
			Scopes:       c.OAuth.Facebook.Scopes,
			Endpoint:     facebook.Endpoint,
		},
		/*	Apple: &oauth2.Config{
			ClientID:     c.OAuth.Apple.ClientID,
			ClientSecret: c.OAuth.Apple.ClientSecret,
			RedirectURL:  c.OAuth.Apple.RedirectURL,
			Scopes:       c.OAuth.Apple.Scopes,
			Endpoint:     apple.Endpoint,
		},*/
	}
}

type OAuthConfig struct {
	Google   *oauth2.Config
	Facebook *oauth2.Config
	//Apple    *oauth2.Config
}
