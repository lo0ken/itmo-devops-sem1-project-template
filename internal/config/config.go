package config

import "os"

// DBConfig представляет конфигурацию базы данных
type DBConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

// ServerConfig представляет конфигурацию сервера
type ServerConfig struct {
	Port string
}

// Config представляет общую конфигурацию приложения
type Config struct {
	DB     DBConfig
	Server ServerConfig
}

// LoadConfig загружает конфигурацию из переменных окружения
// Использует значения по умолчанию если переменные не заданы
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

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
