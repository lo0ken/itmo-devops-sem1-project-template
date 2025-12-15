package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"project_sem/internal/models"
	"project_sem/internal/repository"
	"project_sem/internal/services"
)

// PricesHandler обрабатывает HTTP запросы к /api/v0/prices
type PricesHandler struct {
	archiveService   *services.ArchiveService
	csvService       *services.CSVService
	validatorService *services.ValidatorService
	repo             *repository.PriceRepository
}

// NewPricesHandler создает новый экземпляр handler
func NewPricesHandler(
	archiveService *services.ArchiveService,
	csvService *services.CSVService,
	validatorService *services.ValidatorService,
	repo *repository.PriceRepository,
) *PricesHandler {
	return &PricesHandler{
		archiveService:   archiveService,
		csvService:       csvService,
		validatorService: validatorService,
		repo:             repo,
	}
}

// HandlePost обрабатывает POST /api/v0/prices
func (h *PricesHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	// Установка Content-Type для JSON ответа
	w.Header().Set("Content-Type", "application/json")

	// Шаг 1: Парсинг query параметра "type" (default: "zip")
	archiveType := r.URL.Query().Get("type")
	if archiveType == "" {
		archiveType = "zip"
	}

	// Проверка типа архива
	if archiveType != "zip" && archiveType != "tar" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid archive type"})
		return
	}

	// Шаг 2: Парсинг multipart form, извлечение файла
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to parse form"})
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Printf("Failed to get file from form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()

	// Чтение файла в память
	fileData, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to read file"})
		return
	}

	// Шаг 3: Извлечение CSV из архива
	csvData, err := h.archiveService.Extract(fileData, archiveType)
	if err != nil {
		log.Printf("Failed to extract archive: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "corrupted archive"})
		return
	}

	// Шаг 4: Парсинг CSV
	rawRecords, totalCount, err := h.csvService.Parse(csvData)
	if err != nil {
		log.Printf("Failed to parse CSV: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid CSV format"})
		return
	}

	// Шаг 5: Валидация и обнаружение дубликатов
	validationResult, err := h.validatorService.Validate(rawRecords, totalCount)
	if err != nil {
		log.Printf("Failed to validate data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	// Шаг 6: Вставка валидных записей в БД
	err = h.repo.BulkInsert(validationResult.ValidRecords)
	if err != nil {
		log.Printf("Failed to insert data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to insert data"})
		return
	}

	// Шаг 7: Получение статистики из БД
	stats, err := h.repo.GetStatistics()
	if err != nil {
		log.Printf("Failed to get statistics: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	// Шаг 8: Формирование ответа
	response := models.UploadResponse{
		TotalCount:      validationResult.TotalCount,
		DuplicatesCount: validationResult.DuplicatesCount,
		TotalItems:      stats.TotalItems,
		TotalCategories: stats.TotalCategories,
		TotalPrice:      stats.TotalPrice,
	}

	// Шаг 9: Отправка JSON ответа
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
