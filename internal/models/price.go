package models

import "time"

// Price представляет запись о цене товара
type Price struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Category   string    `json:"category"`
	Price      float64   `json:"price"`
	CreateDate time.Time `json:"create_date"`
}

// UploadResponse представляет JSON ответ POST /api/v0/prices
type UploadResponse struct {
	TotalCount      int     `json:"total_count"`      // всего строк в исходном CSV файле
	DuplicatesCount int     `json:"duplicates_count"` // количество дубликатов
	TotalItems      int     `json:"total_items"`      // общее количество записей в БД
	TotalCategories int     `json:"total_categories"` // общее количество уникальных категорий в БД
	TotalPrice      float64 `json:"total_price"`      // суммарная стоимость всех товаров в БД
}

// Statistics представляет агрегированные данные из БД
type Statistics struct {
	TotalItems      int     // общее количество записей в БД
	TotalCategories int     // уникальных категорий в БД
	TotalPrice      float64 // суммарная стоимость
}
