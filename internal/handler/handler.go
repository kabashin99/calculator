// CalculateHandler godoc
// @Summary Calculate the result of an expression
// @Description Calculate the result of a mathematical expression
// @Tags calculator
// @Accept json
// @Produce json
// @Param request body models.Request true "Expression to calculate"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 422 {object} models.ErrorResponse
// @Failure 405 {object} models.ErrorResponse
// @Router /api/v1/calculate [post]
package handler

import (
	"calculator_app/internal/calculator"
	"calculator_app/internal/models"
	"calculator_app/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
)

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed) // STATUS 405
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "метод не разрешен"})
		return
	}

	var req models.Request

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // STATUS 400
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "ошибка декодирования JSON"})
		return
	}

	if !utils.IsValidExpression(req.Expression) {
		w.WriteHeader(http.StatusUnprocessableEntity) // STATUS 422
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "недопустимые символы"})
		return
	}

	result, err := calculator.Calc(req.Expression)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)                       // STATUS 422
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()}) //
		return
	}

	response := models.SuccessResponse{Result: fmt.Sprintf("%f", result)}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // STATUS 200
	json.NewEncoder(w).Encode(response)
}
