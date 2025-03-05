package handler

import (
	service "calculator_app/internal/orchestrator/service"
	"encoding/json"
	"net/http"
)

type Handler struct {
	orc *service.Orchestrator // Используем тип из пакета service
}

func NewHandler(orc *service.Orchestrator) *Handler {
	return &Handler{orc: orc}
}

func (h *Handler) AddExpression(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}

	id, err := h.orc.AddExpression(req.Expression) // Добавляем выражение
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *Handler) GetExpressions(w http.ResponseWriter, r *http.Request) {
	expressions := h.orc.GetExpressions() // Получаем все выражения
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": expressions})
}

func (h *Handler) GetExpressionByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	expr, exists := h.orc.GetExpressionByID(id) // Получаем выражение по ID
	if !exists {
		http.Error(w, "expression not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"expression": expr})
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	task, exists := h.orc.GetTask() // Получаем задачу
	if !exists {
		http.Error(w, "no tasks available", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"task": task})
}

func (h *Handler) SubmitResult(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}

	if !h.orc.SubmitResult(req.ID, req.Result) { // Отправляем результат
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
