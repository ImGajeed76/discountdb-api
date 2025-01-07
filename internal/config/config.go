package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	REDISHost     string
	REDISPort     string
	REDISUser     string
	REDISPassword string

	DeepSeekAPIKey string

	GMailUser   string
	GMailAPIKey string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USERNAME"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_DATABASE"),
		DBSSLMode:  os.Getenv("DB_SSL_MODE"),

		REDISHost:     os.Getenv("REDIS_HOST"),
		REDISPort:     os.Getenv("REDIS_PORT"),
		REDISUser:     os.Getenv("REDIS_USERNAME"),
		REDISPassword: os.Getenv("REDIS_PASSWORD"),

		DeepSeekAPIKey: os.Getenv("DEEP_SEEK_API_KEY"),

		GMailUser:   os.Getenv("GMAIL_USER"),
		GMailAPIKey: os.Getenv("GMAIL_API_KEY"),
	}

	return config, nil
}
