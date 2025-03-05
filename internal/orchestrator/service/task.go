package service

import (
	"calculator_app/internal/orchestrator/models"
	"calculator_app/internal/orchestrator/repository"
	"log"
)

type TaskService struct {
	taskRepo repository.TaskRepository
	exprRepo repository.ExpressionRepository
}

func NewTaskService(taskRepo repository.TaskRepository, exprRepo repository.ExpressionRepository) *TaskService {
	return &TaskService{
		taskRepo: taskRepo,
		exprRepo: exprRepo,
	}
}

func (s *TaskService) GetNextTask() (*models.Task, error) {
	return s.taskRepo.GetNextPending()
}

func (s *TaskService) SubmitResult(taskID string, result float64) error {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return err
	}

	task.Result = result
	task.Status = "done"

	log.Printf("📌 Задача %s завершена с результатом %f", task.ID, result)

	// Теперь обновляем не только статус, но и результат в БД
	if err := s.taskRepo.UpdateStatus(taskID, "done", result); err != nil {
		return err
	}

	// Проверяем, все ли задачи у выражения завершены
	expressionTasks, err := s.taskRepo.GetByExpressionID(task.ExpressionID)
	if err != nil {
		return err
	}

	allDone := true
	var finalResult float64 = 0
	for _, t := range expressionTasks {
		log.Printf("🔍 Проверяем задачу %s (статус: %s, результат: %f)", t.ID, t.Status, t.Result)
		if t.Status != "done" {
			allDone = false
			break
		}
		finalResult += t.Result
	}

	if allDone {
		log.Printf("✅ Все задачи выражения %s завершены! Итог: %f", task.ExpressionID, finalResult)
		if err := s.exprRepo.UpdateResult(task.ExpressionID, finalResult, "done"); err != nil {
			return err
		}
	}

	return nil
}

func (s *TaskService) GetTask(taskID string) (*models.Task, error) {
	return s.taskRepo.GetByID(taskID)
}

func (s *TaskService) GetAllTasks() ([]*models.Task, error) {
	return s.taskRepo.GetAll()
}

func (s *TaskService) IsTaskReady(task *models.Task) bool {
	for _, depID := range task.DependsOn {
		depTask, err := s.taskRepo.GetByID(depID)
		if err != nil || depTask.Status != "done" {
			return false
		}
	}
	return true
}
