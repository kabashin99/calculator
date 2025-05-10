package main

import (
	"calculator_app/internal/agent"
	"calculator_app/internal/config"
	"log"
)

func main() {
	cfg, err := config.LoadConfig("config/config.txt")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Agent started with %d workers", cfg.ComputingPower)
	agentInstance := agent.NewAgent("localhost:50051", cfg.ComputingPower)
	log.Println("gRPC агент запущен на порту 50051")
	agentInstance.Start()

	select {}
}
