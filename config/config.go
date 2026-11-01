package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Postgres PostgresConfig
	JWT      JWTConfig
	Log      LogConfig
	Storage  StorageConfig
}

type (
	AppConfig struct {
		Name            string
		Port            string
		ShutdownTimeout time.Duration
	}

	StorageConfig struct {
		Endpoint      string
		Region        string
		Bucket        string
		AccessKey     string
		SecretKey     string
		ForcePathStyle bool
	}

	PostgresConfig struct {
		Host     string
		Port     string
		User     string
		Password string
		Database string
		SSLMode  string
	}

	JWTConfig struct {
		Secret string
		Issuer string
		TTL    time.Duration
	}

	LogConfig struct {
		Level string
	}
)

func Load() Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not loaded: %v", err)
	}

	return Config{
		App: AppConfig{
			Name:            env("APP_NAME", "stitchforce-server"),
			Port:            env("APP_PORT", "8080"),
			ShutdownTimeout: durationEnv("APP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Postgres: PostgresConfig{
			Host:     env("POSTGRES_HOST", "localhost"),
			Port:     env("POSTGRES_PORT", "5432"),
			User:     env("POSTGRES_USER", "postgres"),
			Password: env("POSTGRES_PASSWORD", "postgres"),
			Database: env("POSTGRES_DB", "stitchforce"),
			SSLMode:  env("POSTGRES_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret: env("JWT_SECRET", "change-me-in-production"),
			Issuer: env("JWT_ISSUER", "stitchforce-server"),
			TTL:    durationEnv("JWT_TTL", 24*time.Hour),
		},
		Log: LogConfig{
			Level: env("LOG_LEVEL", "info"),
		},
		Storage: StorageConfig{
			Endpoint:      env("STORAGE_ENDPOINT", ""),
			Region:        env("STORAGE_REGION", ""),
			Bucket:        env("STORAGE_BUCKET", ""),
			AccessKey:     env("STORAGE_ACCESS_KEY", ""),
			SecretKey:     env("STORAGE_SECRET_KEY", ""),
			ForcePathStyle: env("STORAGE_FORCE_PATH_STYLE", "true") == "true",
		},
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return time.Duration(seconds) * time.Second
}
