package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"calculator_app/internal/orchestrator/service"
	"calculator_app/internal/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestAddExpression(t *testing.T) {
	orc := service.NewOrchestrator(100, 100, 200, 200)
	h := NewHandler(orc)

	reqBody := `{"expression": "3+5*2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.AddExpression(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var respBody map[string]string
	json.NewDecoder(res.Body).Decode(&respBody)
	assert.NotEmpty(t, respBody["id"])
}

func TestGetExpressions(t *testing.T) {
	orc := service.NewOrchestrator(100, 100, 200, 200)
	h := NewHandler(orc)

	orc.AddExpression("2+2")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
	w := httptest.NewRecorder()

	h.GetExpressions(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var respBody map[string][]models.Expression
	json.NewDecoder(res.Body).Decode(&respBody)
	assert.Len(t, respBody["expressions"], 1)
}

func TestGetExpressionByID(t *testing.T) {
	orc := service.NewOrchestrator(100, 100, 200, 200)
	h := NewHandler(orc)

	exprID, _ := orc.AddExpression("10 / 2")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions/"+exprID, nil)
	req.SetPathValue("id", exprID)
	w := httptest.NewRecorder()

	h.GetExpressionByID(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var respBody map[string]models.Expression
	json.NewDecoder(res.Body).Decode(&respBody)
	assert.Equal(t, exprID, respBody["expression"].ID)
}

func TestGetExpressionByID_NotFound(t *testing.T) {
	orc := service.NewOrchestrator(100, 100, 200, 200)
	h := NewHandler(orc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions/nonexistent-id", nil)
	req.SetPathValue("id", "nonexistent-id")
	w := httptest.NewRecorder()

	h.GetExpressionByID(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestGetTask(t *testing.T) {
	orc := service.NewOrchestrator(100, 100, 200, 200)
	h := NewHandler(orc)

	orc.AddExpression("3 + 5")

	req := httptest.NewRequest(http.MethodGet, "/internal/task", nil)
	w := httptest.NewRecorder()

	h.GetTask(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var respBody map[string]models.Task
	json.NewDecoder(res.Body).Decode(&respBody)
	assert.NotEmpty(t, respBody["task"].ID)
}

func TestGetTask_NotFound(t *testing.T) {
	orc := service.NewOrchestrator(100, 100, 200, 200)
	h := NewHandler(orc)

	req := httptest.NewRequest(http.MethodGet, "/internal/task", nil)
	w := httptest.NewRecorder()

	h.GetTask(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestSubmitResult(t *testing.T) {
	orc := service.NewOrchestrator(100, 100, 200, 200)
	h := NewHandler(orc)

	exprID, _ := orc.AddExpression("3 + 5")
	task, _ := orc.GetTask()

	reqBody, _ := json.Marshal(map[string]interface{}{
		"id":     task.ID,
		"result": 8.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/internal/task", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.SubmitResult(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Проверяем, обновился ли результат выражения
	expr, _ := orc.GetExpressionByID(exprID)
	assert.Equal(t, 8.0, expr.Result)
	assert.Equal(t, "done", expr.Status)
}

func TestSubmitResult_InvalidRequest(t *testing.T) {
	orc := service.NewOrchestrator(100, 100, 200, 200)
	h := NewHandler(orc)

	req := httptest.NewRequest(http.MethodPost, "/internal/task", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.SubmitResult(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
}

func TestSubmitResult_TaskNotFound(t *testing.T) {
	orc := service.NewOrchestrator(100, 100, 200, 200)
	h := NewHandler(orc)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"id":     "nonexistent-task",
		"result": 8.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/internal/task", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.SubmitResult(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}
