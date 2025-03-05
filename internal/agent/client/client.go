package client

import (
	"bytes"
	"calculator_app/internal/orchestrator/models"
	"encoding/json"
	"net/http"
	"time"
)

type OrchestratorClient interface {
	FetchTask() (*models.TaskResponse, error)
	SendResult(taskID string, result float64) error
	GetTask(taskID string) (*models.Task, error)
	SubmitResult(task *models.Task) error
}

type orchestratorClientImpl struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) OrchestratorClient {
	return &orchestratorClientImpl{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *orchestratorClientImpl) FetchTask() (*models.TaskResponse, error) {
	resp, err := c.client.Get(c.baseURL + "/internal/task")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	var task models.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (c *orchestratorClientImpl) SendResult(taskID string, result float64) error {
	payload := map[string]interface{}{
		"id":     taskID,
		"result": result,
	}

	jsonBody, _ := json.Marshal(payload)
	resp, err := c.client.Post(
		c.baseURL+"/internal/task",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	return nil
}

func (c *orchestratorClientImpl) GetTask(taskID string) (*models.Task, error) {
	resp, err := c.client.Get(c.baseURL + "/tasks/" + taskID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	var task models.Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (c *orchestratorClientImpl) SubmitResult(task *models.Task) error {
	payload, _ := json.Marshal(task)
	req, err := http.NewRequest("POST", c.baseURL+"/internal/task", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
