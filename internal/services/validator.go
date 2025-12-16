package services

import (
	"strconv"
	"strings"
	"time"

	"project_sem/internal/models"
	"project_sem/internal/repository"
)

type ValidatorService struct {
	repo *repository.PriceRepository
}

func NewValidatorService(repo *repository.PriceRepository) *ValidatorService {
	return &ValidatorService{repo: repo}
}

type ValidationResult struct {
	ValidRecords    []models.Price
	TotalCount      int
	DuplicatesCount int
}

func (v *ValidatorService) Validate(rawRecords []RawPriceRecord, totalCount int) (*ValidationResult, error) {
	result := &ValidationResult{
		ValidRecords:    make([]models.Price, 0),
		TotalCount:      totalCount,
		DuplicatesCount: 0,
	}

	seenIDs := make(map[int]bool)
	validIDs := make([]int, 0)
	tempValidRecords := make([]models.Price, 0)

	for _, raw := range rawRecords {
		if strings.TrimSpace(raw.ID) == "" ||
			strings.TrimSpace(raw.Name) == "" ||
			strings.TrimSpace(raw.Category) == "" ||
			strings.TrimSpace(raw.Price) == "" ||
			strings.TrimSpace(raw.CreateDate) == "" {
			continue
		}

		id, err := strconv.Atoi(strings.TrimSpace(raw.ID))
		if err != nil {
			continue
		}

		price, err := strconv.ParseFloat(strings.TrimSpace(raw.Price), 64)
		if err != nil || price < 0 {
			continue
		}

		createDate, err := time.Parse("2006-01-02", strings.TrimSpace(raw.CreateDate))
		if err != nil {
			continue
		}

		if seenIDs[id] {
			result.DuplicatesCount++
			continue
		}

		seenIDs[id] = true
		validIDs = append(validIDs, id)

		validRecord := models.Price{
			ID:         id,
			Name:       strings.TrimSpace(raw.Name),
			Category:   strings.TrimSpace(raw.Category),
			Price:      price,
			CreateDate: createDate,
		}

		tempValidRecords = append(tempValidRecords, validRecord)
	}

	existingMap, err := v.repo.CheckExistingIDs(validIDs)
	if err != nil {
		return nil, err
	}

	for _, record := range tempValidRecords {
		if existingMap[record.ID] {
			result.DuplicatesCount++
		} else {
			result.ValidRecords = append(result.ValidRecords, record)
		}
	}

	return result, nil
}
