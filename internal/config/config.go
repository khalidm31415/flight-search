package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application configuration.
type Config struct {
	AppPort           string
	RedisHost         string
	RedisPort         string
	RedisPassword     string
	RedisDB           int
	CacheTTL          time.Duration
	ProviderTimeout   time.Duration
	AirAsiaMaxRetries int
	AirAsiaBaseDelay  time.Duration
}

// Load reads configuration from environment variables, falling back to defaults.
func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		AppPort:           getEnv("APP_PORT", "3000"),
		RedisHost:         getEnv("REDIS_HOST", "localhost"),
		RedisPort:         getEnv("REDIS_PORT", "6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		RedisDB:           getEnvAsInt("REDIS_DB", 0),
		CacheTTL:          time.Duration(getEnvAsInt("CACHE_TTL_SECONDS", 300)) * time.Second,
		ProviderTimeout:   time.Duration(getEnvAsInt("PROVIDER_TIMEOUT_MS", 3000)) * time.Millisecond,
		AirAsiaMaxRetries: getEnvAsInt("AIRASIA_MAX_RETRIES", 3),
		AirAsiaBaseDelay:  time.Duration(getEnvAsInt("AIRASIA_BASE_DELAY_MS", 100)) * time.Millisecond,
	}
}

// RedisAddr returns the full Redis address.
func (c *Config) RedisAddr() string {
	return c.RedisHost + ":" + c.RedisPort
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strVal := getEnv(key, "")
	if strVal == "" {
		return fallback
	}
	val, err := strconv.Atoi(strVal)
	if err != nil {
		return fallback
	}
	return val
}
