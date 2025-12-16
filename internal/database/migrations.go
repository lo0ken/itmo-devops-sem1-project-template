package database

import (
	"database/sql"
	"fmt"
)

func RunMigrations(db *sql.DB) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS prices (
		id INTEGER PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		category VARCHAR(255) NOT NULL,
		price NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
		create_date DATE NOT NULL
	);
	`

	if _, err := db.Exec(createTableQuery); err != nil {
		return fmt.Errorf("failed to create prices table: %w", err)
	}

	createIndexQuery := `
	CREATE INDEX IF NOT EXISTS idx_prices_category ON prices(category);
	`

	if _, err := db.Exec(createIndexQuery); err != nil {
		return fmt.Errorf("failed to create index on category: %w", err)
	}

	return nil
}
