package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
)

// CSVService предоставляет методы для работы с CSV файлами
type CSVService struct{}

// NewCSVService создает новый экземпляр сервиса
func NewCSVService() *CSVService {
	return &CSVService{}
}

// RawPriceRecord представляет "сырую" запись из CSV (все поля - строки)
type RawPriceRecord struct {
	LineNumber int
	ID         string
	Name       string
	Category   string
	Price      string
	CreateDate string
}

// Parse парсит CSV данные и возвращает массив "сырых" записей
// Возвращает также totalCount - количество строк данных (без заголовка)
func (s *CSVService) Parse(data []byte) ([]RawPriceRecord, int, error) {
	reader := csv.NewReader(bytes.NewReader(data))

	// Чтение заголовка (первая строка)
	header, err := reader.Read()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Проверка наличия всех необходимых колонок
	if len(header) < 5 {
		return nil, 0, fmt.Errorf("invalid CSV format: expected at least 5 columns, got %d", len(header))
	}

	var records []RawPriceRecord
	lineNumber := 1 // Начинаем с 1 (после заголовка)

	// Чтение строк данных
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read CSV row at line %d: %w", lineNumber+1, err)
		}

		lineNumber++

		// Проверка количества колонок
		if len(row) < 5 {
			// Пропускаем строки с недостаточным количеством колонок
			continue
		}

		record := RawPriceRecord{
			LineNumber: lineNumber,
			ID:         row[0],
			Name:       row[1],
			Category:   row[2],
			Price:      row[3],
			CreateDate: row[4],
		}

		records = append(records, record)
	}

	totalCount := len(records)
	return records, totalCount, nil
}
