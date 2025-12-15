package database

import (
	"database/sql"
	"fmt"
)

// RunMigrations выполняет SQL миграции для создания таблиц и индексов
func RunMigrations(db *sql.DB) error {
	// SQL запрос для создания таблицы prices
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS prices (
		id INTEGER PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		category VARCHAR(255) NOT NULL,
		price NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
		create_date DATE NOT NULL
	);
	`

	// Выполнение создания таблицы
	if _, err := db.Exec(createTableQuery); err != nil {
		return fmt.Errorf("failed to create prices table: %w", err)
	}

	// SQL запрос для создания индекса на category
	createIndexQuery := `
	CREATE INDEX IF NOT EXISTS idx_prices_category ON prices(category);
	`

	// Выполнение создания индекса
	if _, err := db.Exec(createIndexQuery); err != nil {
		return fmt.Errorf("failed to create index on category: %w", err)
	}

	return nil
}
