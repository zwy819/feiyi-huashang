package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port               string
	AppEnv             string
	DBPath             string
	JWTSecret          string
	JWTExpiresIn       string
	CORSOrigin         string
	UploadDir          string
	MaxUploadSize      int64
	MaxImageUploadSize int64

	// AI docent. Keys are only read from the environment (or the ignored
	// server/.env file); they are never exposed by the status endpoint.
	AIEnabled          bool
	AIBaseURL          string
	AIAPIKey           string
	AIChatModel        string
	AIVisionModel      string
	AIEmbeddingModel   string
	AIEmbeddingDim     int
	AIDailyTokenBudget int64
	AIRateLimitPerMin  int
	AIMaxHistoryTurns  int
	AIMockMode         bool
	AIRequestTimeout   time.Duration
}

func Load() *Config {
	loadDotEnv()

	return &Config{
		Port:               getEnv("PORT", "3001"),
		AppEnv:             getEnv("APP_ENV", "development"),
		DBPath:             getEnv("DB_PATH", "./data/nonlegacy.db"),
		JWTSecret:          getEnv("JWT_SECRET", "nonlegacy_dev_secret_key"),
		JWTExpiresIn:       getEnv("JWT_EXPIRES_IN", "168h"),
		CORSOrigin:         getEnv("CORS_ORIGIN", "http://localhost:5173"),
		UploadDir:          getEnv("UPLOAD_DIR", "./data/models"),
		MaxUploadSize:      getEnvInt64("MAX_UPLOAD_SIZE", 52428800),
		MaxImageUploadSize: getEnvInt64("MAX_IMAGE_UPLOAD_SIZE", 10485760),

		AIEnabled:          getEnvBool("AI_ENABLED", true),
		AIBaseURL:          strings.TrimRight(getEnv("AI_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1"), "/"),
		AIAPIKey:           strings.TrimSpace(getEnv("AI_API_KEY", "")),
		AIChatModel:        getEnv("AI_CHAT_MODEL", "qwen-plus"),
		AIVisionModel:      getEnv("AI_VISION_MODEL", "qwen-vl-max"),
		AIEmbeddingModel:   getEnv("AI_EMBEDDING_MODEL", "text-embedding-v3"),
		AIEmbeddingDim:     getEnvInt("AI_EMBEDDING_DIM", 1024),
		AIDailyTokenBudget: getEnvInt64("AI_DAILY_TOKEN_BUDGET", 2000000),
		AIRateLimitPerMin:  getEnvInt("AI_RATE_LIMIT_PER_MIN", 10),
		AIMaxHistoryTurns:  getEnvInt("AI_MAX_HISTORY_TURNS", 8),
		AIMockMode:         getEnvBool("AI_MOCK_MODE", false),
		AIRequestTimeout:   getEnvDuration("AI_REQUEST_TIMEOUT", 60*time.Second),
	}
}

func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

// loadDotEnv reads .env without overriding variables that are already set by
// the process. It deliberately ignores malformed lines instead of failing the
// server: a missing AI key should degrade to browsing, not prevent startup.
func loadDotEnv() {
	for _, path := range []string{".env", filepath.Join("server", ".env")} {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key = strings.TrimSpace(strings.TrimPrefix(key, "\ufeff"))
			if key == "" {
				continue
			}
			value = strings.TrimSpace(value)
			if len(value) >= 2 {
				if (value[0] == '"' && value[len(value)-1] == '"') ||
					(value[0] == '\'' && value[len(value)-1] == '\'') {
					value = value[1 : len(value)-1]
				}
			}
			if _, exists := os.LookupEnv(key); exists {
				continue
			}
			_ = os.Setenv(key, value)
		}
		_ = f.Close()
	}
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

func getEnvInt(key string, defaultVal int) int {
	return int(getEnvInt64(key, int64(defaultVal)))
}

func getEnvBool(key string, defaultVal bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}
