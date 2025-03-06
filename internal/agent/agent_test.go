package agent

import (
	"calculator_app/internal/pkg/models"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchTask(t *testing.T) {
	expectedTask := &models.Task{
		ID:        "task-1",
		Arg1:      3,
		Arg2:      5,
		Operation: "+",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/task" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]*models.Task{"task": expectedTask})
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	agent := NewAgent(server.URL, 1)
	task, err := agent.fetchTask()

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, expectedTask.ID, task.ID)
	assert.Equal(t, expectedTask.Operation, task.Operation)
}

func TestExecuteTask(t *testing.T) {
	agent := NewAgent("http://localhost", 1)

	tests := []struct {
		name     string
		task     *models.Task
		expected float64
	}{
		{"Addition", &models.Task{Arg1: 3, Arg2: 5, Operation: "+"}, 8},
		{"Subtraction", &models.Task{Arg1: 10, Arg2: 3, Operation: "-"}, 7},
		{"Multiplication", &models.Task{Arg1: 4, Arg2: 6, Operation: "*"}, 24},
		{"Division", &models.Task{Arg1: 10, Arg2: 2, Operation: "/"}, 5},
		{"Division by zero", &models.Task{Arg1: 10, Arg2: 0, Operation: "/"}, 0},
		{"Unknown operation", &models.Task{Arg1: 10, Arg2: 5, Operation: "^"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.executeTask(tt.task)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSubmitResult(t *testing.T) {
	expectedTaskID := "task-1"
	expectedResult := 42.0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/task" && r.Method == http.MethodPost {
			body, _ := ioutil.ReadAll(r.Body)
			defer r.Body.Close()

			var payload map[string]interface{}
			json.Unmarshal(body, &payload)

			if payload["id"] == expectedTaskID && payload["result"] == expectedResult {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	agent := NewAgent(server.URL, 1)
	err := agent.submitResult(expectedTaskID, expectedResult)

	assert.NoError(t, err)
}

func TestFetchTaskError(t *testing.T) {
	// Запускаем сервер, который всегда возвращает 404
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	agent := NewAgent(server.URL, 1)
	task, err := agent.fetchTask()

	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestSubmitResultError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	agent := NewAgent(server.URL, 1)
	err := agent.submitResult("task-1", 42.0)

	assert.Error(t, err)
}
