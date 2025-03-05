package worker

import (
	//"context"
	"calculator_app/config"
	agentClient "calculator_app/internal/agent/client"
	"calculator_app/internal/orchestrator/models"
	"errors"
	"fmt"
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

func (w *Worker) StartWorker() {
	for {
		log.Println("🔄 Запрашиваю задачу у оркестратора...")
		task, err := w.client.FetchTask()
		if err != nil {
			log.Printf("❌ Ошибка при получении задачи: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if task == nil {
			log.Println("⚠️ Нет доступных задач. Ожидаю...")
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf("✅ Получена задача: %s %s %s", task.Arg1, task.Operation, task.Arg2)

		result, err := w.executeTask(task)
		if err != nil {
			log.Printf("❌ Ошибка при выполнении задачи: %v", err)
			continue
		}

		log.Printf("📤 Отправляю результат: %f", result)
		if err := w.client.SendResult(task.ID, result); err != nil {
			log.Printf("❌ Ошибка при отправке результата: %v", err)
		} else {
			log.Println("✅ Результат успешно отправлен!")
		}
	}
}

func StartWorkers(client agentClient.OrchestratorClient, config *config.Config, numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		worker := NewWorker(client, config)
		go worker.StartWorker()
	}
}

func (w *Worker) executeTask(task *models.TaskResponse) (float64, error) {
	if task.Operation == "" {
		log.Printf("Task %s has empty operation", task.ID)
		return 0, errors.New("task has empty operation")
	}

	// Преобразуем аргументы в числа, если они являются ID задач
	arg1, err := parseOperand(task.Arg1)
	if err != nil {
		return 0, fmt.Errorf("invalid arg1: %v", err)
	}

	arg2, err := parseOperand(task.Arg2)
	if err != nil {
		return 0, fmt.Errorf("invalid arg2: %v", err)
	}

	// Имитация длительной операции
	opTime := w.getOperationTime(task.Operation)
	time.Sleep(opTime)

	// Выполняем операцию
	switch task.Operation {
	case "+":
		return arg1 + arg2, nil
	case "-":
		return arg1 - arg2, nil
	case "*":
		return arg1 * arg2, nil
	case "/":
		if arg2 == 0 {
			return 0, errors.New("division by zero")
		}
		return arg1 / arg2, nil
	default:
		return 0, errors.New("unknown operation")
	}
}

func parseOperand(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		// Логика получения результата задачи по ID
		return 0, errors.New("task dependencies not implemented")
	default:
		return 0, fmt.Errorf("unsupported type: %T", val)
	}
}

func (w *Worker) getOperationTime(operation string) time.Duration {
	switch operation {
	case "+":
		return w.config.AddTime
	case "-":
		return w.config.SubTime
	case "*":
		return w.config.MulTime
	case "/":
		return w.config.DivTime
	default:
		return 1 * time.Second
	}
}

func (w *Worker) resolveOperand(operand interface{}) (float64, error) {
	switch v := operand.(type) {
	case float64:
		return v, nil // Если это число — просто возвращаем

	case string:
		// Если передали UUID задачи, ожидаем её завершения
		log.Printf("⏳ Ожидание результата для задачи %s...", v)
		for {
			task, err := w.client.GetTask(v)
			if err != nil {
				log.Printf("❌ Ошибка при получении задачи %s: %v", v, err)
				time.Sleep(1 * time.Second)
				continue
			}

			if task.Status == "done" {
				log.Printf("✅ Задача %s завершена, результат: %f", v, task.Result)
				return task.Result, nil
			}

			if task.Status == "error" {
				return 0, errors.New("зависимая задача завершилась с ошибкой")
			}

			time.Sleep(1 * time.Second) // Ждём завершения задачи
		}

	default:
		return 0, fmt.Errorf("неизвестный тип операнда: %T", operand)
	}
}
