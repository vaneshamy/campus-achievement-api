package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv memuat environment variables dari file .env
func LoadEnv() {
	// Cek apakah file .env ada
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		log.Println("Warning: .env file not found, using system environment variables")
		return
	}

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	log.Println("Environment variables loaded successfully")
}

// GetEnv mendapatkan environment variable dengan default value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}