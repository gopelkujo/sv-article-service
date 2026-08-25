// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds runtime configuration for the article service.
type Config struct {
	AppEnv              string
	AppPort             string
	CORSAllowedOrigins  []string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBMaxOpenConns      int
	DBMaxIdleConns      int
	DBConnMaxLifetime   time.Duration
	HTTPReadTimeout     time.Duration
	HTTPWriteTimeout    time.Duration
	HTTPIdleTimeout     time.Duration
	HTTPShutdownTimeout time.Duration
}

// Load reads configuration from the environment.
// When a .env file is present it is loaded for local development; missing
// .env is ignored so production can rely solely on process env vars.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		AppPort:            getEnv("APP_PORT", "8080"),
		CORSAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
		DBHost:             getEnv("DB_HOST", "127.0.0.1"),
		DBPort:             getEnv("DB_PORT", "3306"),
		DBUser:             getEnv("DB_USER", "article"),
		DBPassword:         getEnv("DB_PASSWORD", "article"),
		DBName:             getEnv("DB_NAME", "article"),
	}

	var err error

	cfg.DBMaxOpenConns, err = getEnvInt("DB_MAX_OPEN_CONNS", 25)
	if err != nil {
		return nil, fmt.Errorf("config: DB_MAX_OPEN_CONNS: %w", err)
	}

	cfg.DBMaxIdleConns, err = getEnvInt("DB_MAX_IDLE_CONNS", 5)
	if err != nil {
		return nil, fmt.Errorf("config: DB_MAX_IDLE_CONNS: %w", err)
	}

	cfg.DBConnMaxLifetime, err = getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("config: DB_CONN_MAX_LIFETIME: %w", err)
	}

	cfg.HTTPReadTimeout, err = getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("config: HTTP_READ_TIMEOUT: %w", err)
	}

	cfg.HTTPWriteTimeout, err = getEnvDuration("HTTP_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("config: HTTP_WRITE_TIMEOUT: %w", err)
	}

	cfg.HTTPIdleTimeout, err = getEnvDuration("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return nil, fmt.Errorf("config: HTTP_IDLE_TIMEOUT: %w", err)
	}

	cfg.HTTPShutdownTimeout, err = getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("config: HTTP_SHUTDOWN_TIMEOUT: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// DSN returns a MySQL data source name suitable for database/sql.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

// Addr returns the HTTP listen address (host:port).
func (c *Config) Addr() string {
	return ":" + c.AppPort
}

func (c *Config) validate() error {
	if c.AppPort == "" {
		return fmt.Errorf("config: APP_PORT is required")
	}
	if c.DBHost == "" {
		return fmt.Errorf("config: DB_HOST is required")
	}
	if c.DBName == "" {
		return fmt.Errorf("config: DB_NAME is required")
	}
	if c.DBUser == "" {
		return fmt.Errorf("config: DB_USER is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q: %w", raw, err)
	}
	return value, nil
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", raw, err)
	}
	return value, nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
