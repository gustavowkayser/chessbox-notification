package application

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	Secret      string
	AppMode     Mode
}

func LoadConfig() Config {
	// Attempt to load .env file for local development; ignore if missing (e.g. in AWS Lambda)
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mode, _ := DetectMode(os.Getenv("APP_MODE"))

	return Config{
		Port:        port,
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Secret:      os.Getenv("SECRET"),
		AppMode:     mode,
	}
}
