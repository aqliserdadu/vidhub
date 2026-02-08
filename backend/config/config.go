package config

import (
	"os"
	"strconv"
	"strings"

	"videodownload/internal/model"

	"github.com/joho/godotenv"
)

// Load loads configuration from environment variables
func Load() *model.Config {
	godotenv.Load()

	return &model.Config{
		Server: model.ServerConfig{
			Port:    getEnvInt("SERVER_PORT", 8080),
			Host:    getEnvStr("SERVER_HOST", "0.0.0.0"),
			Timeout: getEnvInt("SERVER_TIMEOUT", 300),
		},
		Storage: model.StorageConfig{
			DownloadDir:     getEnvStr("DOWNLOAD_DIR", "./downloads"),
			MaxVideoSizeMB:  getEnvInt("MAX_VIDEO_SIZE_MB", 300),
			CleanupInterval: getEnvInt("STORAGE_CLEANUP_INTERVAL", 3600),
			FileTTLSeconds:  getEnvInt("FILE_TTL_SECONDS", 86400),
		},
		Python: model.PythonConfig{
			Port:    getEnvInt("PYTHON_WORKER_PORT", 5000),
			Host:    getEnvStr("PYTHON_WORKER_HOST", "localhost"),
			Timeout: getEnvInt("PYTHON_WORKER_TIMEOUT", 60),
		},
		Logging: model.LoggingConfig{
			Level:        getEnvStr("LOG_LEVEL", "info"),
			FilePath:     getEnvStr("LOG_FILE", "./log/app.log"),
			RotationSize: getEnvInt64("LOG_ROTATION_SIZE", 104857600),
			MaxBackups:   getEnvInt("LOG_MAX_BACKUPS", 3),
			MaxAge:       getEnvInt("LOG_MAX_AGE", 7),
		},
		Security: model.SecurityConfig{
			AllowedDomains: strings.Split(getEnvStr("ALLOWED_DOMAINS", "youtube.com,youtu.be,vimeo.com,facebook.com,m.facebook.com,fb.watch,tiktok.com,instagram.com,twitter.com,x.com"), ","),
			RequestTimeout: getEnvInt("REQUEST_TIMEOUT", 60),
			RateLimitPerIP: getEnvInt("RATE_LIMIT_PER_IP", 30),
		},
		Quota: model.QuotaConfig{
			Enabled:      getEnvBool("QUOTA_ENABLED", false),
			DailyLimitMB: getEnvInt64("QUOTA_DAILY_LIMIT_MB", 1000),
			ResetHour:    getEnvInt("QUOTA_RESET_HOUR", 0),
			ResetMinute:  getEnvInt("QUOTA_RESET_MINUTE", 0),
		},
		RateLimit: model.RateLimitConfig{
			Enabled:           getEnvBool("RATELIMIT_ENABLED", true),
			RequestsPerMinute: getEnvInt("RATELIMIT_REQUESTS_PER_MINUTE", 60),
			BurstSize:         getEnvInt("RATELIMIT_BURST_SIZE", 10),
			CleanupInterval:   getEnvInt("RATELIMIT_CLEANUP_INTERVAL", 1800),
		},
		QualityCategories: model.QualityCategoriesConfig{
			Enabled: parseEnabledQualityCategories(
				getEnvStr("ENABLED_QUALITY_CATEGORIES", "Audio,FD,SD,HD,FHD"),
			),
		},
	}
}

// parseEnabledQualityCategories parses comma-separated quality categories from env
func parseEnabledQualityCategories(categoriesStr string) []string {
	if categoriesStr == "" {
		// Default: all categories enabled
		return []string{"Audio", "FD", "SD", "HD", "FHD"}
	}

	categories := strings.Split(categoriesStr, ",")
	var validCategories []string

	validCategoryMap := map[string]bool{
		"Audio": true,
		"FD":    true,
		"SD":    true,
		"HD":    true,
		"FHD":   true,
	}

	for _, cat := range categories {
		cat = strings.TrimSpace(cat)
		if validCategoryMap[cat] {
			validCategories = append(validCategories, cat)
		}
	}

	// If no valid categories specified, use default
	if len(validCategories) == 0 {
		return []string{"Audio", "FD", "SD", "HD", "FHD"}
	}

	return validCategories
}

func getEnvStr(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	valStr := getEnvStr(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvInt64(key string, defaultVal int64) int64 {
	valStr := getEnvStr(key, "")
	if val, err := strconv.ParseInt(valStr, 10, 64); err == nil {
		return val
	}
	return defaultVal
}
func getEnvBool(key string, defaultVal bool) bool {
	valStr := strings.ToLower(getEnvStr(key, ""))
	if valStr == "true" || valStr == "1" || valStr == "yes" {
		return true
	}
	if valStr == "false" || valStr == "0" || valStr == "no" {
		return false
	}
	return defaultVal
}
