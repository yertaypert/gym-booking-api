package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	ServerPort string
	JWTTTL     time.Duration
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "gym_booking"),
		JWTSecret:  getEnv("JWT_SECRET", "super-secret-change-me"),
		ServerPort: normalizeServerPort(getEnv("SERVER_PORT", ":8080")),
		JWTTTL:     24 * time.Hour,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func normalizeServerPort(port string) string {
	if strings.Contains(port, ":") {
		return port
	}

	return ":" + port
}
