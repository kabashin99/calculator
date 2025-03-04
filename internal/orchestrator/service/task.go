package service

import (
	"calculator_app/internal/orchestrator/models"
	"calculator_app/internal/orchestrator/repository"
)

type TaskService struct {
	taskRepo *repository.TaskRepository
}

func NewTaskService(taskRepo *repository.TaskRepository) *TaskService {
	return &TaskService{
		taskRepo: taskRepo,
	}
}

func (s *TaskService) GetNextTask() (*models.Task, error) {
	return s.taskRepo.GetNextPending()
}

func (s *TaskService) SubmitResult(taskID string, result float64) error {
	// Update task status
	if err := s.taskRepo.UpdateStatus(taskID, "completed"); err != nil {
		return err
	}

	// Update related expression
	// (Implementation depends on your task-expression relationship)
	return nil
}
