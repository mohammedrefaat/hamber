package config

type Config struct {
	Database  DatabaseConfig  `yaml:"database"`
	Server    ServerConfig    `yaml:"server"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	OAuth     OAuthConfigs    `yaml:"oauth"`
	JWT       JWTConfig       `yaml:"jwt"`
	Email     EmailConfig     `yaml:"email"`
	Storage   StorageConfig   `yaml:"storage"`
	Payment   PaymentConfig   `yaml:"payment"`
	RabbitMQ  RabbitMQConfig  `yaml:"rabbitmq"`
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

type PaymentConfig struct {
	Fawry  FawryConfig  `yaml:"fawry"`
	Paymob PaymobConfig `yaml:"paymob"`
}

type FawryConfig struct {
	MerchantCode string `yaml:"merchant_code"`
	SecurityKey  string `yaml:"security_key"`
	APIURL       string `yaml:"api_url"`
	CallbackURL  string `yaml:"callback_url"`
	Enabled      bool   `yaml:"enabled"`
}

type PaymobConfig struct {
	APIKey        string `yaml:"api_key"`
	IntegrationID string `yaml:"integration_id"`
	IframeID      string `yaml:"iframe_id"`
	HMACSecret    string `yaml:"hmac_secret"`
	APIURL        string `yaml:"api_url"`
	CallbackURL   string `yaml:"callback_url"`
	Enabled       bool   `yaml:"enabled"`
}

type RabbitMQConfig struct {
	URL      string `yaml:"url"`
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Vhost    string `yaml:"vhost"`
}
