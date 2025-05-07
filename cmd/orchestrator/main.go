package main

import (
	"calculator_app/db"
	"calculator_app/internal/config"
	"calculator_app/internal/orchestrator/handler"
	"calculator_app/internal/orchestrator/repository"
	"calculator_app/internal/orchestrator/service"
	"database/sql"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
)

func main() {
	cfg, err := config.LoadConfig("config/config.txt")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	os.MkdirAll("db", os.ModePerm)
	connStr := "file:db/calculator.db?cache=shared&mode=rwc"
	dbConn, err := sql.Open("sqlite", connStr)
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}
	defer dbConn.Close()

	if err := dbConn.Ping(); err != nil {
		log.Fatalf("DB Ping failed: %v", err)
	}

	if err := db.RunMigrations(dbConn); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	repo := repository.NewRepository(dbConn)
	orc := service.NewOrchestrator(cfg.TimeAdditionMS, cfg.TimeSubtractionMS, cfg.TimeMultiplicationMS, cfg.TimeDivisionMS, repo)
	OrchHandler := handler.NewHandler(orc)

	http.HandleFunc("POST /api/v1/register", OrchHandler.RegisterUser)
	http.HandleFunc("POST /api/v1/login", OrchHandler.LoginUser)
	http.HandleFunc("POST /api/v1/calculate", OrchHandler.AddExpression)
	http.HandleFunc("GET /api/v1/expressions", OrchHandler.GetExpressions)
	http.HandleFunc("GET /api/v1/expressions/{id}", OrchHandler.GetExpressionByID)

	http.HandleFunc("GET /internal/task", OrchHandler.GetTask)
	http.HandleFunc("POST /internal/task", OrchHandler.SubmitResult)
	http.HandleFunc("GET /internal/task/{id}", OrchHandler.GetTaskResult)

	log.Println("Оркестратор запущен на localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
