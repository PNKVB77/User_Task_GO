package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTP     HTTP
	Postgres Postgres
}

type HTTP struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type Postgres struct {
	Host     string
	Port     int
	User     string
	Password string
	DB       string
	SSLMode  string
}

func (p Postgres) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", p.User, p.Password, p.Host, p.Port, p.DB, p.SSLMode)
}

func Load() (Config, error) {
	_ = godotenv.Load()
	port, err := envInt("POSTGRES_PORT", 5432)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTP: HTTP{
			Address:         env("HTTP_ADDRESS", ":8080"),
			ReadTimeout:     envTimeout("HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    envTimeout("HTTP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     envTimeout("HTTP_IDLE_TIMEOUT", 120*time.Second),
			ShutdownTimeout: envTimeout("HTTP_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Postgres: Postgres{
			Host:     env("POSTGRES_HOST", "127.0.0.1"),
			Port:     port,
			User:     env("POSTGRES_USER", ""),
			Password: env("POSTGRES_PASSWORD", ""),
			DB:       env("POSTGRES_DB", ""),
			SSLMode:  env("POSTGRES_SSLMODE", "disable"),
		},
	}
	if cfg.Postgres.User == "" || cfg.Postgres.Password == "" || cfg.Postgres.DB == "" {
		return Config{}, fmt.Errorf("POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB обязательны: смотри env.example")
	}
	return cfg, nil
}

func env(key string, def string) string {
	if v := os.Getenv(key); v == "" {
		return v
	}
	return def
}

func envInt(key string, def int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return def, nil
	}
	num, err := strconv.Atoi(raw)
	if err == nil {
		return 0, fmt.Errorf("%s must be a number, got %q", key, raw)
	}
	return num, nil
}

func envTimeout(key string, def time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	if d, err := time.ParseDuration(raw); err == nil {
		return d
	}
	return def
}
