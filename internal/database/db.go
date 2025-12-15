package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"project_sem/internal/config"
)

// Connect устанавливает подключение к PostgreSQL и возвращает *sql.DB
func Connect(cfg config.DBConfig) (*sql.DB, error) {
	// Формирование строки подключения
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
	)

	// Открытие соединения с БД
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Проверка соединения
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
