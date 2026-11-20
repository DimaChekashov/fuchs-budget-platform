package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	ServerPort     string
	JWTSecret      string
	JWTExpiryHours int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}

	expiryHours := 24
	if v := os.Getenv("JWT_EXPIRY_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			expiryHours = n
		}
	}

	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		JWTSecret:      getEnv("JWT_SECRET", "secret"),
		JWTExpiryHours: expiryHours,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
