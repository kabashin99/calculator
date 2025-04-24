package service

import (
	"calculator_app/internal/config"
	"calculator_app/internal/orchestrator/repository"
	"database/sql"
)

func InitOrchestrator(cfg *config.Config, db *sql.DB) *Orchestrator {
	repo := repository.NewRepository(db)
	return NewOrchestrator(
		cfg.TimeAdditionMS,
		cfg.TimeSubtractionMS,
		cfg.TimeMultiplicationMS,
		cfg.TimeDivisionMS,
		repo,
	)
}
