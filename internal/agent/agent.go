package agent

import (
	"calculator_app/internal/pkg/models"
	pb "calculator_app/internal/proto"
	"context"
	"database/sql"
	"errors"
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
			continue
		}

		result, err := a.executeTask(task)
		if err != nil {
			log.Printf("Task %s failed: %v", task.ID, err)

			var taskErr *models.TaskError
			if errors.As(err, &taskErr) {
				if subErr := a.submitWithRetry(task.ID, nil, 3, taskErr); subErr != nil {
					log.Printf("Failed to submit error for task %s: %v", task.ID, subErr)
				}
			} else {
				internalErr := models.NewTaskError(models.ErrInternalError, err.Error())
				_ = a.submitWithRetry(task.ID, nil, 3, internalErr)
			}
			continue
		}

		if err := a.submitWithRetry(task.ID, &result, 3, nil); err != nil {
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

func (a *Agent) submitWithRetry(taskID string, result *float64, maxRetries int, taskErr *models.TaskError) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		var err error
		if taskErr != nil {
			err = a.submitError(taskID, taskErr)
		} else {
			err = a.submitResult(taskID, result)
		}

		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(time.Second * time.Duration(i+1))
	}
	return fmt.Errorf("after %d attempts: %w", maxRetries, lastErr)
}

func (a *Agent) fetchTask() (*models.Task, error) {
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

func (a *Agent) executeTask(task *models.Task) (float64, error) {
	log.Printf("Executing task: %s %f %s %f", task.Operation, task.Arg1, task.Operation, task.Arg2)
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2, nil
	case "-":
		return task.Arg1 - task.Arg2, nil
	case "*":
		return task.Arg1 * task.Arg2, nil
	case "/":
		if task.Arg2 == 0 {
			log.Printf("Division by zero in task ID: %s", task.ID)
			return 0, models.NewTaskError(models.ErrDivisionByZero, "division by zero")
		}
		return task.Arg1 / task.Arg2, nil
	default:
		log.Printf("Unknown operation: %s in task ID: %s", task.Operation, task.ID)
		return 0, models.NewTaskError(models.ErrUnknownOperation, "unknown operation")
	}
}

func (a *Agent) submitResult(taskID string, result *float64) error {
	_, err := a.client.SubmitResult(context.Background(), &pb.SubmitResultRequest{
		TaskId: taskID,
		Outcome: &pb.SubmitResultRequest_Result{
			Result: *result,
		},
	})
	return err
}

func (a *Agent) submitError(taskID string, taskErr *models.TaskError) error {
	_, err := a.client.SubmitResult(context.Background(), &pb.SubmitResultRequest{
		TaskId: taskID,
		Outcome: &pb.SubmitResultRequest_Error{
			Error: string(taskErr.Code),
		},
	})
	return err
}

func (a *Agent) getDependencyResult(taskID string) (float64, error) {
	resp, err := a.client.GetTaskResult(context.Background(), &pb.GetTaskResultRequest{TaskId: taskID})
	if err != nil || !resp.TaskExists {
		return 0, fmt.Errorf("result not available")
	}

	if resp.Result != nil {
		return resp.Result.GetValue(), nil // Извлекаем значение из DoubleValue
	}

	return 0, fmt.Errorf("result not available")
}
