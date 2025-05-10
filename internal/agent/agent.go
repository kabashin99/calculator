package agent

import (
	"calculator_app/internal/pkg/models"
	pb "calculator_app/internal/proto"
	"context"
	"database/sql"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

type Agent struct {
	orchestratorURL string
	computingPower  int
	db              *sql.DB
	client          pb.OrchestratorServiceClient
}

func NewAgent(orchestratorURL string, computingPower int) *Agent {
	conn, err := grpc.NewClient(
		orchestratorURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect to gRPC server: %v", err)
	}

	client := pb.NewOrchestratorServiceClient(conn)

	return &Agent{
		orchestratorURL: orchestratorURL,
		computingPower:  computingPower,
		client:          client,
	}
}

func (a *Agent) Start() {
	// log.Printf("Agent started with %d workers", a.computingPower)

	for i := 0; i < a.computingPower; i++ {
		go a.worker()
	}
}

func (a *Agent) worker() {
	for {
		task, err := a.fetchTask()
		if err != nil {
			log.Printf("Failed to fetch task: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		//log.Printf("агент получил таску %+v", task)

		for _, depID := range task.DependsOn {
			for attempt := 0; attempt < 10; attempt++ {
				result, err := a.getDependencyResult(depID)
				if err == nil {
					if task.Arg1 == 0 {
						task.Arg1 = result
					} else {
						task.Arg2 = result
					}
					break
				}

				if attempt == 9 {
					log.Printf("Dependency %s not ready after 10 attempts", depID)
					continue
				}
				time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
			}
		}
		if task.ID == "" || task.Operation == "" {
			//log.Printf("❗ Пропущена пустая или некорректная задача: %+v", task)
			continue
		}
		log.Printf("Получил таску: %+v ", task)
		result := a.executeTask(task)

		if err := a.submitWithRetry(task.ID, result, 3); err != nil {
			log.Printf("Failed to submit result for task %s: %v", task.ID, err)
		}
	}
}

func (a *Agent) waitForDependencies(task *models.Task) error {
	resolved := 0
	for _, depID := range task.DependsOn {
		for attempt := 0; attempt < 10; attempt++ {
			result, err := a.getDependencyResult(depID)
			if err == nil {
				if resolved == 0 {
					task.Arg1 = result
				} else {
					task.Arg2 = result
				}
				resolved++
				break
			}
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
		}
	}
	return nil
}

func (a *Agent) submitWithRetry(taskID string, result float64, maxRetries int) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		err := a.submitResult(taskID, result)
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(time.Second * time.Duration(i+1))
	}
	return fmt.Errorf("after %d attempts: %w", maxRetries, lastErr)
}

func (a *Agent) fetchTask() (*models.Task, error) {
	/*
		resp, err := http.Get(a.orchestratorURL + "/internal/task")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch task: %v", err)
		}
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				log.Printf("Warning: failed to close response body: %v", cerr)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("no tasks available, status code: %d", resp.StatusCode)
		}

		var response struct {
			Task *models.Task `json:"task"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode task: %v", err)
		}

		if response.Task == nil {
			return nil, fmt.Errorf("empty task response")
		}

		return response.Task, nil
	*/
	resp, err := a.client.GetTask(context.Background(), &pb.GetTaskRequest{})
	if err != nil {
		return nil, err
	}

	return &models.Task{
		ID:            resp.TaskId,
		Operation:     resp.Operation,
		Arg1:          resp.Arg1,
		Arg2:          resp.Arg2,
		OperationTime: int(resp.OperationTime),
		DependsOn:     resp.DependsOn,
		UserLogin:     resp.UserLogin,
	}, nil
}

func (a *Agent) executeTask(task *models.Task) float64 {
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
	/*
		payload := map[string]interface{}{
			"id":     taskID,
			"result": result,
		}
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %v", err)
		}

		log.Printf("Submitting result: Task ID=%s, Result=%f", taskID, result) // ✅ Лог

		resp, err := http.Post(a.orchestratorURL+"/internal/task", "application/json", bytes.NewReader(jsonData))
		if err != nil {
			return fmt.Errorf("failed to submit result: %v", err)
		}
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				log.Printf("Warning: failed to close response body: %v", cerr)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to submit result: status %d", resp.StatusCode)
		}

		log.Printf("Result submitted successfully: Task ID=%s", taskID) // ✅ Лог
		return nil
	*/
	_, err := a.client.SubmitResult(context.Background(), &pb.SubmitResultRequest{
		TaskId: taskID,
		Result: float32(result),
	})
	return err
}

func (a *Agent) getDependencyResult(taskID string) (float64, error) {
	/*
		resp, err := http.Get(a.orchestratorURL + "/internal/task/" + taskID)
		if err != nil {
			return 0, fmt.Errorf("failed to fetch dependency: %w", err)
		}
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				log.Printf("Warning: failed to close response body: %v", cerr)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("dependency not ready")
		}

		var response struct {
			Result float64 `json:"result"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return 0, fmt.Errorf("failed to decode response: %w", err)
		}

		return response.Result, nil
	*/
	resp, err := a.client.GetTaskResult(context.Background(), &pb.GetTaskResultRequest{TaskId: taskID})
	if err != nil || !resp.TaskExists {
		return 0, fmt.Errorf("result not available")
	}

	if resp.Result != nil {
		return resp.Result.GetValue(), nil // Извлекаем значение из DoubleValue
	}

	return 0, fmt.Errorf("result not available")
}
