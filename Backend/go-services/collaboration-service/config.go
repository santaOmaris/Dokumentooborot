package main

import (
	"fmt"
	"os"
)

type config struct {
	DBHost        string
	DBPort        string
	DBName        string
	DBUser        string
	DBPassword    string
	HTTPPort      string
	AmqpURL       string
	AllowedOrigin string
}

func loadConfig() config {
	return config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBName:        getEnv("DB_NAME", "docflow"),
		DBUser:        getEnv("DB_USER", "collaboration_user"),
		DBPassword:    getEnv("DB_PASSWORD", "collaboration_pass"),
		HTTPPort:      getEnv("HTTP_PORT", "8083"),
		AmqpURL:       getEnv("AMQP_URL", "amqp://guest:guest@localhost:5672/"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:3000"),
	}
}

func (c config) dsn() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s search_path=collaboration_schema sslmode=disable",
		c.DBHost, c.DBPort, c.DBName, c.DBUser, c.DBPassword,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
