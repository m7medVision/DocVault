// Package config handles configuration for the worker service.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the worker service.
type Config struct {
	// Environment
	Environment string

	// RabbitMQ
	RabbitMQURL     string
	ReminderQueue   string
	DeadLetterQueue string
	PrefetchCount   int

	// Database
	DatabaseURL string

	// Redis (for distributed locking)
	RedisURL string

	// Reminder extraction
	ReminderExtractionEnabled bool
	DateParseFormats          []string

	// Notification
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	EmailFrom    string

	// Scheduling
	NotifyDaysBefore     []int
	NotifyOnDate         bool
	NotifyAtTime         string // HH:MM format
	MaxRetryAttempts     int
	RetryBackoffDuration time.Duration

	// Logging
	LogLevel string

	// OpenTelemetry
	OTELEndpoint string

	// Sentry
	SentryDSN string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		Environment:               getEnv("ENVIRONMENT", "development"),
		RabbitMQURL:               normalizeRabbitMQURL(getEnv("RABBITMQ_URL", "amqp://docvault:changeme@localhost:5672//docvault")),
		ReminderQueue:             getEnv("RABBITMQ_QUEUE_REMINDER", "docvault.reminder.jobs"),
		DeadLetterQueue:           getEnv("DLQ_QUEUE", "docvault.reminder.jobs.dlq"),
		PrefetchCount:             getEnvInt("PREFETCH_COUNT", 10),
		DatabaseURL:               getEnv("DATABASE_URL", "postgres://docvault:docvault_dev@localhost:5432/docvault"),
		RedisURL:                  getEnv("REDIS_URL", "redis://localhost:6379/0"),
		ReminderExtractionEnabled: getEnvBool("REMINDER_EXTRACTION_ENABLED", true),
		NotifyDaysBefore:          []int{30, 7, 1},
		NotifyOnDate:              true,
		NotifyAtTime:              "09:00",
		MaxRetryAttempts:          getEnvInt("MAX_RETRY_ATTEMPTS", 3),
		RetryBackoffDuration:      5 * time.Minute,
		LogLevel:                  getEnv("LOG_LEVEL", "info"),
		OTELEndpoint:              getEnv("OTEL_EXPORTER_ENDPOINT", ""),
		SentryDSN:                 getEnv("SENTRY_DSN_WORKER", ""),
	}

	// Date formats for parsing
	cfg.DateParseFormats = []string{
		time.RFC3339,
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"2006/01/02",
		"01/02/2006",
		"January 2, 2006",
		"2 January 2006",
		"Jan 2, 2006",
		"2 Jan 2006",
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
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

// getEnv gets an environment variable or returns a default.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvBool gets a boolean environment variable.
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.RabbitMQURL == "" {
		return fmt.Errorf("RABBITMQ_URL is required")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	return nil
}
