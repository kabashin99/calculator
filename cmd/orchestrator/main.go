package main

import (
	"calculator_app/internal/config"
	"calculator_app/internal/orchestrator/handler"
	"calculator_app/internal/orchestrator/service"
	"log"
	"net/http"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("config/config.txt")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем оркестратор с настройками из конфигурации
	orc := service.NewOrchestrator(cfg.TimeAdditionMS, cfg.TimeSubtractionMS, cfg.TimeMultiplicationMS, cfg.TimeDivisionMS)
	OrchHandler := handler.NewHandler(orc)

	// Регистрируем HTTP-обработчики
	http.HandleFunc("POST /api/v1/calculate", OrchHandler.AddExpression)
	http.HandleFunc("GET /api/v1/expressions", OrchHandler.GetExpressions)
	http.HandleFunc("GET /api/v1/expressions/{id}", OrchHandler.GetExpressionByID)
	http.HandleFunc("GET /internal/task", OrchHandler.GetTask)
	http.HandleFunc("POST /internal/task", OrchHandler.SubmitResult)

	log.Println("Оркестратор запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
