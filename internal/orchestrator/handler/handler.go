package handler

import (
	"calculator_app/internal/orchestrator/service"
	"calculator_app/internal/pkg/models"
	"encoding/json"
	"net/http"
)

type Handler struct {
	orc *service.Orchestrator
}

func NewHandler(orc *service.Orchestrator) *Handler {
	return &Handler{orc: orc}
}

func (h *Handler) AddExpression(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusUnprocessableEntity) // 422
		return
	}

	id, err := h.orc.AddExpression(req.Expression)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity) //422
		return
	}

	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *Handler) GetExpressions(w http.ResponseWriter, r *http.Request) {
	exprMap := h.orc.GetExpressions()

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
	expr, exists := h.orc.GetExpressionByID(id)
	if !exists {
		http.Error(w, "expression not found", http.StatusNotFound) //404
		return
	}

	w.WriteHeader(http.StatusOK) // 200
	json.NewEncoder(w).Encode(map[string]interface{}{"expression": expr})
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	task, exists := h.orc.GetTask()
	if !exists {
		http.Error(w, "no tasks available", http.StatusNotFound) //404
		return
	}

	w.WriteHeader(http.StatusOK) // 200
	json.NewEncoder(w).Encode(map[string]interface{}{"task": task})
}

func (h *Handler) SubmitResult(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusUnprocessableEntity) // 422
		return
	}

	if !h.orc.SubmitResult(req.ID, req.Result) {
		http.Error(w, "task not found", http.StatusNotFound) // 404
		return
	}

	w.WriteHeader(http.StatusOK) // 200
}
