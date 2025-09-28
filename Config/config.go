package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
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
	JWT       JWTConfig       `yaml:"jwt"`
	Email     EmailConfig     `yaml:"email"`
	Storage   StorageConfig   `yaml:"storage"`
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
	ReadTimeout    string `yaml:"read_timeout"`
	WriteTimeout   string `yaml:"write_timeout"`
	MaxHeaderBytes int    `yaml:"max_header_bytes"`
}

type RateLimitConfig struct {
	Requests int    `yaml:"requests"`
	Window   string `yaml:"window"`
}

type JWTConfig struct {
	Secret          string `yaml:"secret"`
	EncryptionKey   string `yaml:"encryption_key"` // Base64 encoded AES 32-byte key
	Encrypted       bool   `yaml:"encrypted"`      // NEW: Encryption flag
	ExpirationHours int    `yaml:"expiration_hours"`
}

type EmailConfig struct {
	SMTPHost     string `yaml:"smtp_host"`
	SMTPPort     int    `yaml:"smtp_port"`
	SMTPUsername string `yaml:"smtp_username"`
	SMTPPassword string `yaml:"smtp_password"`
	FromEmail    string `yaml:"from_email"`
	FromName     string `yaml:"from_name"`
}

type StorageConfig struct {
	Type      string      `yaml:"type"`
	LocalPath string      `yaml:"local_path"`
	MinIO     MinIOConfig `yaml:"minio"`
}

type MinIOConfig struct {
	Endpoint           string          `yaml:"endpoint"`
	AccessKey          string          `yaml:"access_key"`
	SecretKey          string          `yaml:"secret_key"`
	UseSSL             bool            `yaml:"use_ssl"`
	Bucket             string          `yaml:"bucket"`
	PublicRead         bool            `yaml:"public_read"`
	MaxFileSize        int64           `yaml:"max_file_size"`
	AllowedExtensions  []string        `yaml:"allowed_extensions"`
	PhotoQuality       int             `yaml:"photo_quality"`
	GenerateThumbnails bool            `yaml:"generate_thumbnails"`
	ThumbnailSizes     []ThumbnailSize `yaml:"thumbnail_sizes"`
}

type ThumbnailSize struct {
	Width  int    `yaml:"width"`
	Height int    `yaml:"height"`
	Suffix string `yaml:"suffix"`
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
	Enabled      bool     `yaml:"enabled"`
}

// Global config instance
var globalConfig *Config
var configFilename string

// AES Encryption Functions
func encryptAES(plaintext string, base64Key string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return "", fmt.Errorf("failed to decode encryption key: %w", err)
	}

	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptAES(ciphertextBase64 string, base64Key string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return "", fmt.Errorf("failed to decode encryption key: %w", err)
	}

	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

// Generate a new 32-byte AES key
func generateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("failed to generate encryption key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	configFilename = filename // Store for later use

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", filename, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml config: %w", err)
	}

	// Generate encryption key if missing
	if cfg.JWT.EncryptionKey == "" {
		key, err := generateEncryptionKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate encryption key: %w", err)
		}
		cfg.JWT.EncryptionKey = key
		fmt.Println("🔑 Generated new AES encryption key for JWT secret")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	cfg.applyDefaults()

	// Handle JWT secret encryption
	if err := cfg.handleJWTEncryption(); err != nil {
		return nil, fmt.Errorf("failed to handle JWT encryption: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Handle JWT secret encryption/decryption
func (c *Config) handleJWTEncryption() error {
	if !c.JWT.Encrypted {
		// Encrypt the secret and save config
		fmt.Println("🔒 Encrypting JWT secret...")

		encryptedSecret, err := encryptAES(c.JWT.Secret, c.JWT.EncryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt JWT secret: %w", err)
		}

		c.JWT.Secret = encryptedSecret
		c.JWT.Encrypted = true

		// Save the updated config back to file
		if err := c.saveConfig(); err != nil {
			return fmt.Errorf("failed to save encrypted config: %w", err)
		}

		fmt.Println("✅ JWT secret encrypted and saved to config file")
	} else {
		fmt.Println("🔓 JWT secret is already encrypted")
	}

	return nil
}

// Save config back to file
func (c *Config) saveConfig() error {
	configData, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create backup of original config
	backupName := configFilename + ".backup"
	if _, err := os.Stat(configFilename); err == nil {
		if err := os.Rename(configFilename, backupName); err != nil {
			fmt.Printf("⚠️  Warning: failed to create backup: %v\n", err)
		} else {
			fmt.Printf("📦 Created backup: %s\n", backupName)
		}
	}

	err = os.WriteFile(configFilename, configData, 0o644)
	if err != nil {
		// Restore backup if save failed
		if _, backupErr := os.Stat(backupName); backupErr == nil {
			os.Rename(backupName, configFilename)
		}
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfig returns the global config instance
func GetConfig() *Config {
	return globalConfig
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
		return 5 * time.Second
	}
	duration, err := time.ParseDuration(c.Server.ReadTimeout)
	if err != nil {
		return 5 * time.Second
	}
	return duration
}

// GetWriteTimeout returns the write timeout as time.Duration
func (c *Config) GetWriteTimeout() time.Duration {
	if c.Server.WriteTimeout == "" {
		return 10 * time.Second
	}
	duration, err := time.ParseDuration(c.Server.WriteTimeout)
	if err != nil {
		return 10 * time.Second
	}
	return duration
}

// GetRateWindow returns the rate limit window as time.Duration
func (c *Config) GetRateWindow() time.Duration {
	if c.RateLimit.Window == "" {
		return time.Minute
	}
	duration, err := time.ParseDuration(c.RateLimit.Window)
	if err != nil {
		return time.Minute
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
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret must not be empty")
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
	if c.JWT.ExpirationHours == 0 {
		c.JWT.ExpirationHours = 24
	}

	// Apply OAuth defaults
	if c.OAuth.Google.Scopes == nil {
		c.OAuth.Google.Scopes = []string{"openid", "profile", "email"}
	}
	if c.OAuth.Facebook.Scopes == nil {
		c.OAuth.Facebook.Scopes = []string{"public_profile", "email"}
	}
	/*if c.OAuth.Apple.Scopes == nil {
		c.OAuth.Apple.Scopes = []string{"name", "email"}
	}*/

	// Storage defaults
	if c.Storage.Type == "" {
		c.Storage.Type = "local"
	}
	if c.Storage.LocalPath == "" {
		c.Storage.LocalPath = "./uploads"
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
		/*Apple: &oauth2.Config{
			ClientID:     c.OAuth.Apple.ClientID,
			ClientSecret: c.OAuth.Apple.ClientSecret,
			RedirectURL:  c.OAuth.Apple.RedirectURL,
			Scopes:       c.OAuth.Apple.Scopes,
			Endpoint:     apple.Endpoint,
		},*/
	}
}

// InitOAuthConfig initializes OAuth configuration from YAML config
func InitOAuthConfig() *OAuthConfig {
	if globalConfig != nil {
		return globalConfig.GetOAuthConfig()
	}

	// Fallback if config not loaded yet
	return &OAuthConfig{
		Google: &oauth2.Config{
			Scopes:   []string{"openid", "profile", "email"},
			Endpoint: google.Endpoint,
		},
		Facebook: &oauth2.Config{
			Scopes:   []string{"public_profile", "email"},
			Endpoint: facebook.Endpoint,
		},
		/*Apple: &oauth2.Config{
			Scopes:   []string{"name", "email"},
			Endpoint: apple.Endpoint,
		},*/
	}
}

type OAuthConfig struct {
	Google   *oauth2.Config
	Facebook *oauth2.Config
	//Apple    *oauth2.Config
}

// JWT helper methods with decryption
func (c *Config) GetJWTSecret() string {
	if c.JWT.Encrypted {
		// Decrypt the secret
		plainSecret, err := decryptAES(c.JWT.Secret, c.JWT.EncryptionKey)
		if err != nil {
			fmt.Printf("⚠️  Error decrypting JWT secret: %v\n", err)
			return "fallback-secret-key" // Fallback
		}
		return plainSecret
	}
	// Return plain secret (shouldn't happen after first run)
	return c.JWT.Secret
}

func (c *Config) GetJWTExpirationHours() int {
	return c.JWT.ExpirationHours
}

// GetStorageType returns the configured storage type
func (c *Config) GetStorageType() string {
	return c.Storage.Type
}

// GetMinIOConfig returns MinIO configuration
func (c *Config) GetMinIOConfig() MinIOConfig {
	return c.Storage.MinIO
}

// IsMinIOEnabled checks if MinIO is the configured storage
func (c *Config) IsMinIOEnabled() bool {
	return strings.ToLower(c.Storage.Type) == "minio"
}

// GetMaxFileSize returns the maximum allowed file size
func (c *Config) GetMaxFileSize() int64 {
	if c.Storage.MinIO.MaxFileSize == 0 {
		return 10485760 // 10MB default (matching your config)
	}
	return c.Storage.MinIO.MaxFileSize
}

// IsFileExtensionAllowed checks if a file extension is allowed
func (c *Config) IsFileExtensionAllowed(ext string) bool {
	if len(c.Storage.MinIO.AllowedExtensions) == 0 {
		// Default allowed extensions
		defaultExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp"}
		for _, allowedExt := range defaultExts {
			if strings.EqualFold(ext, allowedExt) {
				return true
			}
		}
		return false
	}

	for _, allowedExt := range c.Storage.MinIO.AllowedExtensions {
		if strings.EqualFold(ext, allowedExt) {
			return true
		}
	}
	return false
}

// GetPhotoQuality returns the JPEG quality setting
func (c *Config) GetPhotoQuality() int {
	if c.Storage.MinIO.PhotoQuality == 0 {
		return 85 // Default quality
	}
	return c.Storage.MinIO.PhotoQuality
}

// ShouldGenerateThumbnails checks if thumbnail generation is enabled
func (c *Config) ShouldGenerateThumbnails() bool {
	return c.Storage.MinIO.GenerateThumbnails
}

// GetThumbnailSizes returns configured thumbnail sizes
func (c *Config) GetThumbnailSizes() []ThumbnailSize {
	if len(c.Storage.MinIO.ThumbnailSizes) == 0 && c.Storage.MinIO.GenerateThumbnails {
		// Default thumbnail sizes
		return []ThumbnailSize{
			{Width: 150, Height: 150, Suffix: "_thumb"},
			{Width: 400, Height: 300, Suffix: "_small"},
			{Width: 800, Height: 600, Suffix: "_medium"},
		}
	}
	return c.Storage.MinIO.ThumbnailSizes
}

// GetServerPort returns the server port (handles :8088 format)
func (c *Config) GetServerPort() string {
	if strings.HasPrefix(c.Server.Port, ":") {
		return c.Server.Port
	}
	return ":" + c.Server.Port
}

// Update your existing applyDefaults method to include these
func (c *Config) applyStorageDefaults() {
	// Storage defaults
	if c.Storage.Type == "" {
		c.Storage.Type = "local"
	}
	if c.Storage.LocalPath == "" {
		c.Storage.LocalPath = "./uploads"
	}

	// MinIO defaults
	if c.Storage.MinIO.MaxFileSize == 0 {
		c.Storage.MinIO.MaxFileSize = 10485760 // 10MB to match your config
	}
	if c.Storage.MinIO.PhotoQuality == 0 {
		c.Storage.MinIO.PhotoQuality = 85
	}
	if len(c.Storage.MinIO.AllowedExtensions) == 0 {
		c.Storage.MinIO.AllowedExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp"}
	}
	if len(c.Storage.MinIO.ThumbnailSizes) == 0 && c.Storage.MinIO.GenerateThumbnails {
		c.Storage.MinIO.ThumbnailSizes = []ThumbnailSize{
			{Width: 150, Height: 150, Suffix: "_thumb"},
			{Width: 400, Height: 300, Suffix: "_small"},
			{Width: 800, Height: 600, Suffix: "_medium"},
		}
	}
}

// Add this validation to your existing Validate method
func (c *Config) validateStorage() error {
	if c.Storage.Type == "minio" {
		if c.Storage.MinIO.Endpoint == "" {
			return fmt.Errorf("MinIO endpoint cannot be empty when storage type is minio")
		}
		if c.Storage.MinIO.AccessKey == "" || c.Storage.MinIO.SecretKey == "" {
			return fmt.Errorf("MinIO access_key and secret_key cannot be empty")
		}
		if c.Storage.MinIO.Bucket == "" {
			return fmt.Errorf("MinIO bucket name cannot be empty")
		}
	}
	return nil
}
