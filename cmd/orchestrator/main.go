package main

import (
	"calculator_app/internal/config"
	"calculator_app/internal/orchestrator/handler"
	"calculator_app/internal/orchestrator/repository"
	"calculator_app/internal/orchestrator/service"
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.LoadConfig("config/config.txt")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("ошибка подключения к базе: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("база не отвечает: %v", err)
	}
	fmt.Println("✅ Подключение к базе успешно")

	repo := repository.NewRepository(db)

	orc := service.NewOrchestrator(cfg.TimeAdditionMS, cfg.TimeSubtractionMS, cfg.TimeMultiplicationMS, cfg.TimeDivisionMS)
	OrchHandler := handler.NewHandler(orc, repo)

	http.HandleFunc("POST /api/v1/calculate", OrchHandler.AddExpression)
	http.HandleFunc("GET /api/v1/expressions", OrchHandler.GetExpressions)
	http.HandleFunc("GET /api/v1/expressions/{id}", OrchHandler.GetExpressionByID)
	http.HandleFunc("GET /internal/task", OrchHandler.GetTask)
	http.HandleFunc("POST /internal/task", OrchHandler.SubmitResult)

	log.Println("Оркестратор запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
