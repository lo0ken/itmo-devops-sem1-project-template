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
	// Шаг 1: Загрузка конфигурации из переменных окружения
	cfg := config.LoadConfig()
	log.Printf("Configuration loaded: DB=%s:%s/%s, Server=:%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.Database, cfg.Server.Port)

	// Шаг 2: Подключение к PostgreSQL
	db, err := database.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Successfully connected to PostgreSQL")

	// Шаг 3: Запуск миграций (создание таблицы и индексов)
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	// Шаг 4: Инициализация repository, services, handlers
	priceRepo := repository.NewPriceRepository(db)
	archiveService := services.NewArchiveService()
	csvService := services.NewCSVService()
	validatorService := services.NewValidatorService(priceRepo)
	pricesHandler := handlers.NewPricesHandler(archiveService, csvService, validatorService, priceRepo)

	// Шаг 5: Создание роутера (gorilla/mux)
	router := mux.NewRouter()

	// Шаг 6: Регистрация endpoints
	router.HandleFunc("/api/v0/prices", pricesHandler.HandlePost).Methods("POST")
	router.HandleFunc("/api/v0/prices", pricesHandler.HandleGet).Methods("GET")

	// Шаг 7: Запуск HTTP сервера
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
