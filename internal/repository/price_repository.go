package repository

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"project_sem/internal/models"
)

type PriceRepository struct {
	db *sql.DB
}

func NewPriceRepository(db *sql.DB) *PriceRepository {
	return &PriceRepository{db: db}
}

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

func (r *PriceRepository) BulkInsert(prices []models.Price) error {
	if len(prices) == 0 {
		return nil
	}

	query := "INSERT INTO prices (id, name, category, price, create_date) VALUES ($1, $2, $3, $4, $5)"

	for _, p := range prices {
		_, err := r.db.Exec(query, p.ID, p.Name, p.Category, p.Price, p.CreateDate)
		if err != nil {
			continue
		}
	}

	return nil
}

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

type PriceFilter struct {
	StartDate *string
	EndDate   *string
	MinPrice  *float64
	MaxPrice  *float64
}

func (r *PriceRepository) GetFilteredPrices(filter PriceFilter) ([]models.Price, error) {
	query := "SELECT id, name, category, price, create_date FROM prices WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND create_date >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND create_date <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	if filter.MinPrice != nil {
		query += fmt.Sprintf(" AND price >= $%d", argIndex)
		args = append(args, *filter.MinPrice)
		argIndex++
	}

	if filter.MaxPrice != nil {
		query += fmt.Sprintf(" AND price <= $%d", argIndex)
		args = append(args, *filter.MaxPrice)
		argIndex++
	}

	query += " ORDER BY id"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query prices: %w", err)
	}
	defer rows.Close()

	var prices []models.Price
	for rows.Next() {
		var p models.Price
		if err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price, &p.CreateDate); err != nil {
			return nil, fmt.Errorf("failed to scan price: %w", err)
		}
		prices = append(prices, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prices: %w", err)
	}

	return prices, nil
}
