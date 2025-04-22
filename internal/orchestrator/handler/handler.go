package handler

import (
	"calculator_app/internal/orchestrator/repository"
	"calculator_app/internal/orchestrator/service"
	"calculator_app/internal/pkg/models"
	"encoding/json"
	"log"
	"net/http"
)

type Handler struct {
	orchestrator *service.Orchestrator
	repo         *repository.Repository
}

func NewHandler(orc *service.Orchestrator, repo *repository.Repository) *Handler {
	return &Handler{
		orchestrator: orc,
		repo:         repo,
	}
}

func (h *Handler) AddExpression(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusUnprocessableEntity) // 422
		return
	}

	userID := "test-user" // временно

	id, err := h.orchestrator.AddExpression(userID, req.Expression)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity) //422
		return
	}

	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *Handler) GetExpressions(w http.ResponseWriter, r *http.Request) {
	exprMap := h.orchestrator.GetExpressions()

	expressions := make([]models.Expression, 0, len(exprMap))
	for _, expr := range exprMap {
		expressions = append(expressions, *expr)
	}

	w.WriteHeader(http.StatusOK) //200
	json.NewEncoder(w).Encode(map[string]interface{}{
		"expressions": expressions,
	})
}

func (h *Handler) GetExpressionByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	expr, exists := h.orchestrator.GetExpressionByID(id)
	if !exists {
		http.Error(w, "expression not found", http.StatusNotFound) //404
		return
	}

	w.WriteHeader(http.StatusOK) // 200
	json.NewEncoder(w).Encode(map[string]interface{}{"expression": expr})
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	task, err := h.repo.GetTask()
	if err != nil {
		http.Error(w, "Задачи не найдены или ошибка базы", http.StatusInternalServerError)
		log.Printf("Ошибка при получении задачи: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *Handler) SubmitResult(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TaskID string  `json:"task_id"`
		Result float64 `json:"result"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Некорректный формат JSON", http.StatusBadRequest)
		return
	}

	err = h.repo.UpdateTaskResult(input.TaskID, input.Result)
	if err != nil {
		http.Error(w, "Не удалось обновить результат задачи", http.StatusInternalServerError)
		log.Printf("Ошибка при обновлении результата: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(`{"status":"updated"}`))
	if err != nil {
		log.Printf("Ошибка при отправке ответа: %v", err)
	}
}

func (h *Handler) AddTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	err = h.repo.AddTask(&task)
	if err != nil {
		http.Error(w, "Ошибка при добавлении задачи", http.StatusInternalServerError)
		log.Printf("Ошибка при вставке задачи: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
