package handler

import (
	"calculator_app/internal/orchestrator/service"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type ExpressionHandler struct {
	exprService *service.ExpressionService
}

func InitExpressionHandlers(router *mux.Router, s *service.ExpressionService) {
	h := &ExpressionHandler{exprService: s}

	router.HandleFunc("/api/v1/calculate", h.AddExpressionHandler).Methods("POST") // Обрабатываем только POST
	router.HandleFunc("/api/v1/expressions", h.GetExpressionsHandler).Methods("GET")
	router.HandleFunc("/api/v1/expressions/{id}", h.GetExpressionByIDHandler).Methods("GET")
}

func (h *ExpressionHandler) AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity) // 422
		return
	}

	id, err := h.exprService.ProcessExpression(req.Expression)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *ExpressionHandler) GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	expressions, err := h.exprService.GetAllExpressions()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 500
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	var response []map[string]interface{}
	for _, expr := range expressions {
		response = append(response, map[string]interface{}{
			"id":     expr.ID,
			"status": expr.Status,
			"result": expr.Result,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": response})
}

func (h *ExpressionHandler) GetExpressionByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exprID := vars["id"]

	expr, err := h.exprService.GetExpressionByID(exprID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	if expr == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "expression not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"expression": map[string]interface{}{
			"id":     expr.ID,
			"status": expr.Status,
			"result": expr.Result,
		},
	})
}
