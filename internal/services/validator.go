package services

import (
	"strconv"
	"strings"
	"time"

	"project_sem/internal/models"
	"project_sem/internal/repository"
)

// ValidatorService предоставляет методы для валидации данных
type ValidatorService struct {
	repo *repository.PriceRepository
}

// NewValidatorService создает новый экземпляр сервиса
func NewValidatorService(repo *repository.PriceRepository) *ValidatorService {
	return &ValidatorService{repo: repo}
}

// ValidationResult представляет результат валидации
type ValidationResult struct {
	ValidRecords    []models.Price // Валидные записи для вставки
	TotalCount      int            // Всего строк в исходном CSV
	DuplicatesCount int            // Количество дубликатов (в файле + с БД)
}

// Validate валидирует сырые записи и возвращает результат
func (v *ValidatorService) Validate(rawRecords []RawPriceRecord, totalCount int) (*ValidationResult, error) {
	result := &ValidationResult{
		ValidRecords:    make([]models.Price, 0),
		TotalCount:      totalCount,
		DuplicatesCount: 0,
	}

	// Map для отслеживания ID внутри файла
	seenIDs := make(map[int]bool)
	validIDs := make([]int, 0)
	tempValidRecords := make([]models.Price, 0)

	// Этап 1: Проверка полноты и формата, обнаружение дубликатов внутри файла
	for _, raw := range rawRecords {
		// Проверка полноты (все поля непустые)
		if strings.TrimSpace(raw.ID) == "" ||
			strings.TrimSpace(raw.Name) == "" ||
			strings.TrimSpace(raw.Category) == "" ||
			strings.TrimSpace(raw.Price) == "" ||
			strings.TrimSpace(raw.CreateDate) == "" {
			// Невалидная запись - пропускаем
			continue
		}

		// Проверка корректности формата: ID -> int
		id, err := strconv.Atoi(strings.TrimSpace(raw.ID))
		if err != nil {
			// Некорректный ID - пропускаем
			continue
		}

		// Проверка корректности формата: Price -> float64 >= 0
		price, err := strconv.ParseFloat(strings.TrimSpace(raw.Price), 64)
		if err != nil || price < 0 {
			// Некорректная цена - пропускаем
			continue
		}

		// Проверка корректности формата: CreateDate -> "2006-01-02"
		createDate, err := time.Parse("2006-01-02", strings.TrimSpace(raw.CreateDate))
		if err != nil {
			// Некорректная дата - пропускаем
			continue
		}

		// Проверка дубликатов внутри файла
		if seenIDs[id] {
			// Дубликат внутри файла - пропускаем, увеличиваем счетчик
			result.DuplicatesCount++
			continue
		}

		// Первое вхождение ID в файле - отмечаем и добавляем
		seenIDs[id] = true
		validIDs = append(validIDs, id)

		// Создаем валидную запись
		validRecord := models.Price{
			ID:         id,
			Name:       strings.TrimSpace(raw.Name),
			Category:   strings.TrimSpace(raw.Category),
			Price:      price,
			CreateDate: createDate,
		}

		tempValidRecords = append(tempValidRecords, validRecord)
	}

	// Этап 2: Проверка дубликатов с БД
	existingMap, err := v.repo.CheckExistingIDs(validIDs)
	if err != nil {
		return nil, err
	}

	// Фильтрация записей, которые уже есть в БД
	for _, record := range tempValidRecords {
		if existingMap[record.ID] {
			// Дубликат с БД - увеличиваем счетчик, не добавляем в result
			result.DuplicatesCount++
		} else {
			// Новая запись - добавляем в финальный результат
			result.ValidRecords = append(result.ValidRecords, record)
		}
	}

	return result, nil
}
