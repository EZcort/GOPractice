package config

import (
	"log"
	"os"
)

type Config struct {
	MongoDB struct {
		URI      string
		Database string
	}
	Server struct {
		Port string
	}
}

func LoadConfig() *Config {
	cfg := &Config{}

	cfg.MongoDB.URI = getEnv("MONGO_URI", "mongodb://localhost:27017")
	cfg.MongoDB.Database = getEnv("MONGO_DATABASE", "DefaultDB")
	cfg.Server.Port = getEnv("SERVER_PORT", "8080")

	log.Printf("Конфиг загружен\n")

	return cfg
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return defaultValue
}
