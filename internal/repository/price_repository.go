package repository

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"project_sem/internal/models"
)

// PriceRepository представляет репозиторий для работы с ценами
type PriceRepository struct {
	db *sql.DB
}

// NewPriceRepository создает новый экземпляр репозитория
func NewPriceRepository(db *sql.DB) *PriceRepository {
	return &PriceRepository{db: db}
}

// CheckExistingIDs проверяет какие ID уже существуют в базе данных
// Возвращает map[id]exists для быстрой проверки
func (r *PriceRepository) CheckExistingIDs(ids []int) (map[int]bool, error) {
	if len(ids) == 0 {
		return make(map[int]bool), nil
	}

	query := "SELECT id FROM prices WHERE id = ANY($1)"
	rows, err := r.db.Query(query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("failed to check existing IDs: %w", err)
	}
	defer rows.Close()

	existingMap := make(map[int]bool)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan existing ID: %w", err)
		}
		existingMap[id] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating existing IDs: %w", err)
	}

	return existingMap, nil
}

// BulkInsert вставляет массив записей в базу данных построчно
// Без использования транзакции (частичная вставка допустима)
func (r *PriceRepository) BulkInsert(prices []models.Price) error {
	if len(prices) == 0 {
		return nil
	}

	query := "INSERT INTO prices (id, name, category, price, create_date) VALUES ($1, $2, $3, $4, $5)"

	for _, p := range prices {
		_, err := r.db.Exec(query, p.ID, p.Name, p.Category, p.Price, p.CreateDate)
		if err != nil {
			// Игнорируем ошибки вставки отдельных записей (например, дубликаты)
			// Продолжаем вставку оставшихся записей
			continue
		}
	}

	return nil
}

// GetStatistics возвращает агрегированную статистику из ВСЕЙ таблицы БД
func (r *PriceRepository) GetStatistics() (*models.Statistics, error) {
	query := `
		SELECT
			COUNT(*) as total_items,
			COUNT(DISTINCT category) as total_categories,
			COALESCE(SUM(price), 0) as total_price
		FROM prices
	`

	var stats models.Statistics
	err := r.db.QueryRow(query).Scan(&stats.TotalItems, &stats.TotalCategories, &stats.TotalPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return &stats, nil
}
