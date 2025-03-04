package service

import (
	"calculator_app/internal/orchestrator/models"
	"calculator_app/internal/orchestrator/repository"
	"calculator_app/internal/pkg/calculator"
)

type ExpressionService struct {
	exprRepo *repository.ExpressionRepository
	taskRepo *repository.TaskRepository
}

func NewExpressionService(exprRepo *repository.ExpressionRepository, taskRepo *repository.TaskRepository) *ExpressionService {
	return &ExpressionService{
		exprRepo: exprRepo,
		taskRepo: taskRepo,
	}
}

func (s *ExpressionService) ProcessExpression(exprString string) (string, error) {
	// Parse expression and create tasks
	expr := &models.Expression{
		ID:     generateUUID(),
		Status: "processing",
	}

	if err := s.exprRepo.Create(expr); err != nil {
		return "", err
	}

	// Use calculator to split expression into tasks
	tasks, err := calculator.ParseToTasks(exprString)
	if err != nil {
		return "", err
	}

	// Save tasks to repository
	for _, task := range tasks {
		task.ExpressionID = expr.ID
		if err := s.taskRepo.Create(task); err != nil {
			return "", err
		}
	}

	return expr.ID, nil
}
