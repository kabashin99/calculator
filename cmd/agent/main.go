package main

import (
	"calculator_app/internal/agent"
	"calculator_app/internal/config"
	"log"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("config/config.txt")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем агента с настройками из конфигурации
	log.Printf("Agent started with %d workers", cfg.ComputingPower)
	agentInstance := agent.NewAgent("http://localhost:8080", cfg.ComputingPower)
	agentInstance.Start()

	// Бесконечный цикл
	select {}
}
