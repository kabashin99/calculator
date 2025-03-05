package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"calculator_app/internal/pkg/models"
)

type Agent struct {
	orchestratorURL string
	computingPower  int
}

func NewAgent(orchestratorURL string, computingPower int) *Agent {
	return &Agent{
		orchestratorURL: orchestratorURL,
		computingPower:  computingPower,
	}
}

func (a *Agent) Start() {
	results := make(map[string]float64) // Хранилище результатов задач
	log.Printf("Agent started with %d workers", a.computingPower)

	for i := 0; i < a.computingPower; i++ {
		go a.worker(results)
	}
}

func (a *Agent) worker(results map[string]float64) {
	for {
		task, err := a.fetchTask()
		if err != nil {
			log.Printf("Failed to fetch task: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf("Task received: %+v", task)
		result := a.executeTask(task, results)
		log.Printf("Task result: %f", result)

		if err := a.submitResult(task.ID, result); err != nil {
			log.Printf("Failed to submit result: %v", err)
		} else {
			log.Printf("Result submitted for task ID: %s", task.ID)
			results[task.ID] = result // Сохраняем результат для использования в зависимых задачах
		}
	}
}

func (a *Agent) fetchTask() (*models.Task, error) {
	resp, err := http.Get(a.orchestratorURL + "/internal/task")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("no tasks available, status code: %d", resp.StatusCode)
	}

	var response struct {
		Task *models.Task `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode task: %v", err)
	}

	return response.Task, nil
}

func (a *Agent) executeTask(task *models.Task, results map[string]float64) float64 {
	// Подставляем реальные результаты для зависимостей
	for _, depID := range task.DependsOn {
		if result, exists := results[depID]; exists {
			if task.Arg1 == 0 {
				task.Arg1 = result
			} else if task.Arg2 == 0 {
				task.Arg2 = result
			}
		}
	}

	log.Printf("Executing task: %s %f %s %f", task.Operation, task.Arg1, task.Operation, task.Arg2)
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2
	case "-":
		return task.Arg1 - task.Arg2
	case "*":
		return task.Arg1 * task.Arg2
	case "/":
		if task.Arg2 == 0 {
			log.Printf("Division by zero in task ID: %s", task.ID)
			return 0
		}
		return task.Arg1 / task.Arg2
	default:
		log.Printf("Unknown operation: %s in task ID: %s", task.Operation, task.ID)
		return 0
	}
}

func (a *Agent) submitResult(taskID string, result float64) error {
	payload := map[string]interface{}{
		"id":     taskID,
		"result": result,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}

	resp, err := http.Post(a.orchestratorURL+"/internal/task", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to submit result: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to submit result: status %d", resp.StatusCode)
	}

	return nil
}
