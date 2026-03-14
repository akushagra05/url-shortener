package config

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Snowflake SnowflakeConfig
	RateLimit RateLimitConfig
	App      AppConfig
	Cache    CacheConfig
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type SnowflakeConfig struct {
	MachineID      int64
	EpochTimestamp int64
}

type RateLimitConfig struct {
	CreateLimit   int
	RedirectLimit int
	Window        int
}

type AppConfig struct {
	BaseURL     string
	FrontendURL string
}

type CacheConfig struct {
	TTLSeconds         int
	ClickBufferTTLSeconds int
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Parse DATABASE_URL if provided (Render format)
	dbConfig := parseDatabaseURL()
	
	// Parse REDIS_URL if provided (Render format)
	redisConfig := parseRedisURL()

	return &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", getEnv("SERVER_PORT", "8080")),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Database: dbConfig,
		Redis: redisConfig,
		Snowflake: SnowflakeConfig{
			MachineID:      getEnvAsInt64("MACHINE_ID", 1),
			EpochTimestamp: getEnvAsInt64("EPOCH_TIMESTAMP", 1704067200000), // 2024-01-01
		},
		RateLimit: RateLimitConfig{
			CreateLimit:   getEnvAsInt("RATE_LIMIT_CREATE", 10),
			RedirectLimit: getEnvAsInt("RATE_LIMIT_REDIRECT", 100),
			Window:        getEnvAsInt("RATE_LIMIT_WINDOW", 60),
		},
		App: AppConfig{
			BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
		},
		Cache: CacheConfig{
			TTLSeconds:         getEnvAsInt("CACHE_TTL_SECONDS", 86400),
			ClickBufferTTLSeconds: getEnvAsInt("CLICK_BUFFER_TTL_SECONDS", 300),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

func (c *Config) GetDSN() string {
	return "host=" + c.Database.Host +
		" port=" + c.Database.Port +
		" user=" + c.Database.User +
		" password=" + c.Database.Password +
		" dbname=" + c.Database.DBName +
		" sslmode=" + c.Database.SSLMode
}

func (c *Config) GetRedisAddr() string {
	return c.Redis.Host + ":" + c.Redis.Port
}

// parseDatabaseURL parses DATABASE_URL (Render format) or falls back to individual env vars
func parseDatabaseURL() DatabaseConfig {
	databaseURL := os.Getenv("DATABASE_URL")
	
	if databaseURL != "" {
		// Parse postgres://user:password@host:port/dbname?sslmode=require
		u, err := url.Parse(databaseURL)
		if err == nil && u.Scheme == "postgres" {
			password, _ := u.User.Password()
			host := u.Hostname()
			port := u.Port()
			if port == "" {
				port = "5432"
			}
			dbName := strings.TrimPrefix(u.Path, "/")
			sslMode := "require"
			if q := u.Query().Get("sslmode"); q != "" {
				sslMode = q
			}
			
			return DatabaseConfig{
				Host:     host,
				Port:     port,
				User:     u.User.Username(),
				Password: password,
				DBName:   dbName,
				SSLMode:  sslMode,
			}
		}
	}
	
	// Fallback to individual environment variables
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "url_shortener"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// parseRedisURL parses REDIS_URL (Render format) or falls back to individual env vars
func parseRedisURL() RedisConfig {
	redisURL := os.Getenv("REDIS_URL")
	
	if redisURL != "" {
		// Parse redis://default:password@host:port
		u, err := url.Parse(redisURL)
		if err == nil && (u.Scheme == "redis" || u.Scheme == "rediss") {
			password, _ := u.User.Password()
			host := u.Hostname()
			port := u.Port()
			if port == "" {
				port = "6379"
			}
			
			return RedisConfig{
				Host:     host,
				Port:     port,
				Password: password,
				DB:       0,
			}
		}
	}
	
	// Fallback to individual environment variables
	return RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getEnvAsInt("REDIS_DB", 0),
	}
}
