package model

// Config holds application configuration
type Config struct {
	Server            ServerConfig
	Storage           StorageConfig
	Python            PythonConfig
	Logging           LoggingConfig
	Security          SecurityConfig
	Quota             QuotaConfig
	RateLimit         RateLimitConfig
	QualityCategories QualityCategoriesConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port    int
	Host    string
	Timeout int // seconds
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	DownloadDir     string
	MaxVideoSizeMB  int
	CleanupInterval int // seconds
	FileTTLSeconds  int // Time to live for downloaded files
}

// PythonConfig holds Python worker configuration
type PythonConfig struct {
	Port    int
	Host    string
	Timeout int // seconds
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level        string
	FilePath     string
	RotationSize int64 // bytes
	MaxBackups   int
	MaxAge       int // days
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	AllowedDomains []string
	RequestTimeout int // seconds
	RateLimitPerIP int
}

// QuotaConfig holds user download quota configuration
type QuotaConfig struct {
	Enabled      bool  // Enable quota limiting
	DailyLimitMB int64 // Daily quota limit in MB per IP
	ResetHour    int   // Hour (0-23) to reset quota (midnight = 0)
	ResetMinute  int   // Minute (0-59) to reset quota
}

// RateLimitConfig holds rate limiting configuration for DDoS protection
type RateLimitConfig struct {
	Enabled           bool // Enable rate limiting
	RequestsPerMinute int  // Max requests per minute per IP
	BurstSize         int  // Max burst size
	CleanupInterval   int  // Interval in seconds to clean up old entries
}

// QualityCategoriesConfig holds quality category filtering configuration
type QualityCategoriesConfig struct {
	Enabled []string // List of enabled quality categories (Audio, FD, SD, HD, FHD)
	// Examples:
	// - []string{"Audio", "FD", "SD", "HD", "FHD"} = All categories enabled (default)
	// - []string{"SD", "HD", "FHD"} = Only SD, HD, FHD (FD disabled)
	// - []string{"HD", "FHD"} = Only high quality (HD and FHD)
}
