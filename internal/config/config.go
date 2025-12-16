package config

import "os"

type DBConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

type ServerConfig struct {
	Port string
}

type Config struct {
	DB     DBConfig
	Server ServerConfig
}

func LoadConfig() *Config {
	return &Config{
		DB: DBConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			Database: getEnv("POSTGRES_DB", "project-sem-1"),
			User:     getEnv("POSTGRES_USER", "validator"),
			Password: getEnv("POSTGRES_PASSWORD", "val1dat0r"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
