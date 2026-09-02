package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port              string
	AppEnv            string
	DBPath            string
	JWTSecret         string
	JWTExpiresIn      string
	CORSOrigin        string
	UploadDir         string
	MaxUploadSize     int64
	MaxImageUploadSize int64
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "3001"),
		AppEnv:        getEnv("APP_ENV", "development"),
		DBPath:        getEnv("DB_PATH", "./data/nonlegacy.db"),
		JWTSecret:     getEnv("JWT_SECRET", "nonlegacy_dev_secret_key"),
		JWTExpiresIn:  getEnv("JWT_EXPIRES_IN", "168h"),
		CORSOrigin:    getEnv("CORS_ORIGIN", "http://localhost:5173"),
		UploadDir:     getEnv("UPLOAD_DIR", "./data/models"),
		MaxUploadSize:      getEnvInt64("MAX_UPLOAD_SIZE", 52428800),
		MaxImageUploadSize: getEnvInt64("MAX_IMAGE_UPLOAD_SIZE", 10485760),
	}
}

func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt64(key string, defaultVal int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return defaultVal
}
