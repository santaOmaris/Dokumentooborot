package main

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost        string
	DBPort        string
	DBName        string
	DBUser        string
	DBPassword    string
	HTTPPort      string
	GRPCPort      string
	AllowedOrigin string
}

func loadConfig() Config {
	return Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBName:        getEnv("DB_NAME", "docflow"),
		DBUser:        getEnv("DB_USER", "iam_user"),
		DBPassword:    getEnv("DB_PASSWORD", "iam_pass"),
		HTTPPort:      getEnv("HTTP_PORT", "8081"),
		GRPCPort:      getEnv("GRPC_PORT", "9081"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:5173"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?search_path=iam_schema&sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
