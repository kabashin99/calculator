package worker

import (
	"calculator_app/config"
	agentClient "calculator_app/internal/agent/client"
	"calculator_app/internal/orchestrator/models"
	"log"
	"time"
)

type Worker struct {
	client agentClient.OrchestratorClient
	config *config.Config
}

func NewWorker(client agentClient.OrchestratorClient, config *config.Config) *Worker {
	return &Worker{
		client: client,
		config: config,
	}
}

// StartWorker запускает одного воркера
func (w *Worker) StartWorker() {
	for {
		task, err := w.client.FetchTask()
		if err != nil {
			log.Printf("Error fetching task: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if task == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		result := w.executeTask(task)
		if err := w.client.SendResult(task.ID, result); err != nil {
			log.Printf("Failed to send result: %v", err)
		}
	}
}

// StartWorkers запускает несколько воркеров
func StartWorkers(client agentClient.OrchestratorClient, config *config.Config, numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		worker := NewWorker(client, config)
		go worker.StartWorker()
	}
}

func (w *Worker) executeTask(task *models.TaskResponse) float64 {
	// Имитация длительной операции
	opTime := time.Duration(task.OperationTime) * time.Millisecond
	time.Sleep(opTime)

	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2
	case "-":
		return task.Arg1 - task.Arg2
	case "*":
		return task.Arg1 * task.Arg2
	case "/":
		if task.Arg2 == 0 {
			log.Panic("division by zero")
		}
		return task.Arg1 / task.Arg2
	default:
		log.Panicf("unknown operation: %s", task.Operation)
		return 0
	}
}
