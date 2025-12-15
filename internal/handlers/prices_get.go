package handlers

import (
	"log"
	"net/http"
	"strconv"

	"project_sem/internal/repository"
)

func (h *PricesHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	var filter repository.PriceFilter

	if start := queryParams.Get("start"); start != "" {
		filter.StartDate = &start
	}

	if end := queryParams.Get("end"); end != "" {
		filter.EndDate = &end
	}

	if minStr := queryParams.Get("min"); minStr != "" {
		minPrice, err := strconv.ParseFloat(minStr, 64)
		if err != nil || minPrice <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid min parameter"))
			return
		}
		filter.MinPrice = &minPrice
	}

	if maxStr := queryParams.Get("max"); maxStr != "" {
		maxPrice, err := strconv.ParseFloat(maxStr, 64)
		if err != nil || maxPrice <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid max parameter"))
			return
		}
		filter.MaxPrice = &maxPrice
	}

	prices, err := h.repo.GetFilteredPrices(filter)
	if err != nil {
		log.Printf("Failed to get filtered prices: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("database error"))
		return
	}

	csvData, err := h.csvService.Generate(prices)
	if err != nil {
		log.Printf("Failed to generate CSV: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to generate CSV"))
		return
	}

	zipData, err := h.archiveService.CreateZip(csvData, "data.csv")
	if err != nil {
		log.Printf("Failed to create ZIP: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to create archive"))
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
	w.WriteHeader(http.StatusOK)
	w.Write(zipData)
}
