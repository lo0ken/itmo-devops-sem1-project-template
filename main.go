package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"project_sem/internal/config"
	"project_sem/internal/database"
	"project_sem/internal/handlers"
	"project_sem/internal/repository"
	"project_sem/internal/services"
)

func main() {
	cfg := config.LoadConfig()
	log.Printf("Configuration loaded: DB=%s:%s/%s, Server=:%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.Database, cfg.Server.Port)

	db, err := database.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Successfully connected to PostgreSQL")

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	priceRepo := repository.NewPriceRepository(db)
	archiveService := services.NewArchiveService()
	csvService := services.NewCSVService()
	validatorService := services.NewValidatorService(priceRepo)
	pricesHandler := handlers.NewPricesHandler(archiveService, csvService, validatorService, priceRepo)

	router := mux.NewRouter()

	router.HandleFunc("/api/v0/prices", pricesHandler.HandlePost).Methods("POST")
	router.HandleFunc("/api/v0/prices", pricesHandler.HandleGet).Methods("GET")

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
