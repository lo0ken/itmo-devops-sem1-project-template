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

type PricesHandler struct {
	archiveService   *services.ArchiveService
	csvService       *services.CSVService
	validatorService *services.ValidatorService
	repo             *repository.PriceRepository
}

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

func (h *PricesHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	archiveType := r.URL.Query().Get("type")
	if archiveType == "" {
		archiveType = "zip"
	}

	if archiveType != "zip" && archiveType != "tar" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid archive type"})
		return
	}

	err := r.ParseMultipartForm(32 << 20)
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

	fileData, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to read file"})
		return
	}

	csvData, err := h.archiveService.Extract(fileData, archiveType)
	if err != nil {
		log.Printf("Failed to extract archive: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "corrupted archive"})
		return
	}

	rawRecords, totalCount, err := h.csvService.Parse(csvData)
	if err != nil {
		log.Printf("Failed to parse CSV: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid CSV format"})
		return
	}

	validationResult, err := h.validatorService.Validate(rawRecords, totalCount)
	if err != nil {
		log.Printf("Failed to validate data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	stats, duplicatesCount, err := h.repo.InsertAndGetStats(validationResult.ValidRecords)
	if err != nil {
		log.Printf("Failed to insert data and get statistics: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	response := models.UploadResponse{
		TotalCount:      validationResult.TotalCount,
		DuplicatesCount: validationResult.DuplicatesCount + duplicatesCount,
		TotalItems:      stats.TotalItems,
		TotalCategories: stats.TotalCategories,
		TotalPrice:      stats.TotalPrice,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
