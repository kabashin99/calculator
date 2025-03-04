package client

import (
	"bytes"
	"calculator_app/internal/orchestrator/models"
	"encoding/json"
	//"io"
	"net/http"
	"time"
)

type OrchestratorClient interface {
	FetchTask() (*models.TaskResponse, error)
	SendResult(taskID string, result float64) error
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
	return err
}
