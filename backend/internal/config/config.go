package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultEmbeddingModel = "openai/text-embedding-3-large"

var approvedEmbeddingModels = []string{defaultEmbeddingModel}

type Config struct {
	Server  ServerConfig
	DB      DBConfig
	Redis   RedisConfig
	Auth    AuthConfig
	Search  SearchConfig
	Storage StorageConfig
	Queue   QueueConfig
	Obs     ObservabilityConfig
	Log     LogConfig

	// Deprecated: use Obs.SentryDSN
	SentryDSN string
	// Deprecated: use Server.Environment
	Environment string
	// Deprecated: use Obs.OTELEndpoint
	OTELEndpoint string
	// Deprecated: use Search.EmbeddingAPIKey
	OpenRouterAPIKey string
	// Deprecated: use Search.EmbeddingModel
	EmbeddingModel string
	// Deprecated: use Search.EmbeddingAPIKey for Mistral too
	MistralAPIKey string
	// Deprecated: use Auth fields
	JWTAccessTokenTTL  time.Duration
	JWTRefreshTokenTTL time.Duration
	// Deprecated: use Storage fields
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool
	// Deprecated: use Queue fields
	RabbitMQURL      string
	RabbitMQQueueOCR string
	// Deprecated: use Auth fields
	JWTIssuer   string
	JWTAudience string
	JWTSecret   string
	// Deprecated: no replacement
	WorkerAPIKey string
}

type ServerConfig struct {
	Port        int
	Environment string
	CORSOrigins []string
}

type DBConfig struct {
	URL             string
	MaxConns        int
	MinConns        int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type AuthConfig struct {
	JWTSecret          string
	JWTAccessTokenTTL  time.Duration
	JWTRefreshTokenTTL time.Duration
	JWTIssuer          string
	JWTAudience        string
}

type SearchConfig struct {
	Enabled           bool
	EmbeddingProvider string
	EmbeddingModel    string
	EmbeddingAPIKey   string
	EmbeddingURL      string
	EmbeddingDim      int
	ChatModel         string
}

type StorageConfig struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	Bucket          string
	UseSSL          bool
	PresignedURLTTL time.Duration
}

type QueueConfig struct {
	URL           string
	OCRQueue      string
	ProcessQueue  string
	ReminderQueue string
}

type ObservabilityConfig struct {
	SentryDSN    string
	OTELEndpoint string
}

type LogConfig struct {
	Level slog.Level
}

func Load() (*Config, error) {
	if err := validateAISettings(); err != nil {
		return nil, fmt.Errorf("AI config validation failed: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:        getEnvInt("PORT", 8080),
			Environment: getEnvString("ENVIRONMENT", "development"),
			CORSOrigins: getEnvStringSlice("CORS_ORIGINS", []string{"*"}),
		},
		DB: DBConfig{
			URL:             getEnvString("DATABASE_URL", ""),
			MaxConns:        getEnvInt("DB_MAX_CONNS", 25),
			MinConns:        getEnvInt("DB_MIN_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", time.Hour),
		},
		Redis: RedisConfig{
			Host:         getEnvString("REDIS_HOST", "localhost"),
			Port:         getEnvString("REDIS_PORT", "6379"),
			Password:     getEnvString("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 0),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
			MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
			DialTimeout:  getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		},
		Auth: AuthConfig{
			JWTSecret:          getEnvString("JWT_SECRET", ""),
			JWTAccessTokenTTL:  getEnvDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
			JWTRefreshTokenTTL: getEnvDuration("JWT_REFRESH_TOKEN_TTL", 720*time.Hour),
			JWTIssuer:          getEnvString("JWT_ISSUER", "docvault.app"),
			JWTAudience:        getEnvString("JWT_AUDIENCE", "docvault-api"),
		},
		Search: SearchConfig{
			Enabled:           getEnvBool("SEARCH_ENABLED", false),
			EmbeddingProvider: getEnvString("EMBEDDING_PROVIDER", "openrouter"),
			EmbeddingModel:    getEnvString("EMBEDDING_MODEL", defaultEmbeddingModel),
			EmbeddingAPIKey:   getEnvString("OPENROUTER_API_KEY", ""),
			EmbeddingURL:      getEnvString("EMBEDDING_URL", ""),
			EmbeddingDim:      getEnvInt("EMBEDDING_DIM", 1024),
			ChatModel:         getEnvString("OPENROUTER_CHAT_MODEL", "google/gemini-2.0-flash-001"),
		},
		Storage: StorageConfig{
			Endpoint:        getEnvString("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:       getEnvString("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey:       getEnvString("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:          getEnvString("MINIO_BUCKET", "docvault-documents"),
			UseSSL:          getEnvBool("MINIO_USE_SSL", false),
			PresignedURLTTL: getEnvDuration("PRESIGNED_URL_TTL", 15*time.Minute),
		},
		Queue: QueueConfig{
			URL:           normalizeRabbitMQURL(getEnvString("RABBITMQ_URL", "amqp://docvault:changeme@localhost:5672//docvault")),
			OCRQueue:      getEnvString("RABBITMQ_QUEUE_OCR", "docvault.ocr.jobs"),
			ProcessQueue:  getEnvString("RABBITMQ_QUEUE_PROCESSING", "docvault.processing.jobs"),
			ReminderQueue: getEnvString("RABBITMQ_QUEUE_REMINDER", "docvault.reminder.jobs"),
		},
		Obs: ObservabilityConfig{
			SentryDSN:    getEnvString("SENTRY_DSN_BACKEND", ""),
			OTELEndpoint: getEnvString("OTEL_EXPORTER_ENDPOINT", ""),
		},
		Log: LogConfig{
			Level: parseLogLevel(getEnvString("LOG_LEVEL", "info")),
		},

		OpenRouterAPIKey: getEnvString("OPENROUTER_API_KEY", ""),
		EmbeddingModel:   getEnvString("EMBEDDING_MODEL", defaultEmbeddingModel),
		MistralAPIKey:    getEnvString("MISTRAL_API_KEY", ""),
		WorkerAPIKey:     getEnvString("WORKER_API_KEY", ""),
	}

	// Backward compatibility: sync top-level deprecated fields from sub-configs
	cfg.Environment = cfg.Server.Environment
	cfg.SentryDSN = cfg.Obs.SentryDSN
	cfg.OTELEndpoint = cfg.Obs.OTELEndpoint
	cfg.Search.EmbeddingAPIKey = cfg.OpenRouterAPIKey
	cfg.MinIOEndpoint = cfg.Storage.Endpoint
	cfg.MinIOAccessKey = cfg.Storage.AccessKey
	cfg.MinIOSecretKey = cfg.Storage.SecretKey
	cfg.MinIOBucket = cfg.Storage.Bucket
	cfg.MinIOUseSSL = cfg.Storage.UseSSL
	cfg.RabbitMQURL = cfg.Queue.URL
	cfg.RabbitMQQueueOCR = cfg.Queue.OCRQueue
	cfg.JWTAccessTokenTTL = cfg.Auth.JWTAccessTokenTTL
	cfg.JWTRefreshTokenTTL = cfg.Auth.JWTRefreshTokenTTL
	cfg.JWTIssuer = cfg.Auth.JWTIssuer
	cfg.JWTAudience = cfg.Auth.JWTAudience
	cfg.JWTSecret = cfg.Auth.JWTSecret

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Server.Environment == "development" {
		return nil
	}

	var missing []string

	if c.DB.URL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.Auth.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if c.Auth.JWTAudience == "" {
		missing = append(missing, "JWT_AUDIENCE")
	}
	if c.Search.EmbeddingAPIKey == "" {
		missing = append(missing, "OPENROUTER_API_KEY")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

func validateAISettings() error {
	if err := validateEmbeddingModelName(getEnvString("EMBEDDING_MODEL", defaultEmbeddingModel), "EMBEDDING_MODEL"); err != nil {
		return err
	}
	return nil
}

func validateEmbeddingModelName(modelName string, settingName string) error {
	normalized := strings.TrimSpace(modelName)
	if normalized == "" {
		return fmt.Errorf("%s must not be empty", settingName)
	}

	for _, approved := range approvedEmbeddingModels {
		if normalized == approved {
			return nil
		}
	}

	return fmt.Errorf("%s must be one of: %s. Received: %q", settingName, strings.Join(approvedEmbeddingModels, ", "), modelName)
}

func normalizeRabbitMQURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	if (parsed.Scheme != "amqp" && parsed.Scheme != "amqps") || !strings.HasPrefix(parsed.Path, "//") {
		return raw
	}

	parsed.RawPath = "/" + url.PathEscape(parsed.Path[1:])
	return parsed.String()
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if dur, err := time.ParseDuration(value); err == nil {
			return dur
		}
	}
	return defaultValue
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
