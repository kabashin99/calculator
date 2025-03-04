package main

import (
	"calculator_app/config"
	agentClient "calculator_app/internal/agent/client"
	"calculator_app/internal/agent/worker"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := config.Load()

	orchestratorURL := os.Getenv("ORCHESTRATOR_URL")
	if orchestratorURL == "" {
		orchestratorURL = "http://localhost:8080"
	}

	workersStr := os.Getenv("COMPUTING_POWER")
	workers, err := strconv.Atoi(workersStr)
	if err != nil {
		log.Fatalf("Invalid COMPUTING_POWER value: %v", err)
	}

	// Создание клиента
	orchestratorClient := agentClient.NewClient(orchestratorURL)

	// Запуск воркеров
	worker.StartWorkers(orchestratorClient, cfg, workers)

	log.Printf("Agent started with %d workers", workers)
	select {} // Бесконечное ожидание
}
