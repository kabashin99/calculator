package handler

import (
	"calculator_app/internal/calculator"
	"encoding/json"
	"fmt"
	"net/http"
)

type request struct {
	Expression string `json:"expression"`
}

type response struct {
	Result string `json:"result"`
}

// @Summary Get a greeting message
// @Description Get a greeting message
// @Tags greetings
// @Accept json
// @Produce json
// @Success 200 {string} string "{"message": "Hello, World!"}"
// @Router /api/v1/calculate [post]
func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	result, err := calculator.Calc(req.Expression)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusUnprocessableEntity)
		return
	}

	resp := response{Result: fmt.Sprintf("%f", result)}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}
