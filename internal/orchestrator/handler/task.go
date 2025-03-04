package handler

import (
	"calculator_app/internal/orchestrator/models"
	"calculator_app/internal/orchestrator/service"
	"encoding/json"
	"net/http"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// GetNextTaskHandler обработчик для получения следующей задачи
// @Summary Получить задачу для выполнения
// @Description Возвращает следующую доступную задачу для вычисления
// @Tags internal
// @Produce json
// @Success 200 {object} models.TaskResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /internal/task [get]
func (h *TaskHandler) GetNextTaskHandler(w http.ResponseWriter, r *http.Request) {
	task, err := h.taskService.GetNextTask()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
		return
	}

	if task == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "no tasks available"})
		return
	}

	response := models.TaskResponse{
		ID:            task.ID,
		Arg1:          task.Arg1,
		Arg2:          task.Arg2,
		Operation:     task.Operation,
		OperationTime: getOperationTime(task.Operation),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SubmitResultHandler обработчик для отправки результата задачи
// @Summary Отправить результат задачи
// @Description Принимает результат выполнения задачи от агента
// @Tags internal
// @Accept json
// @Produce json
// @Param request body models.TaskResult true "Результат задачи"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 422 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /internal/task [post]
func (h *TaskHandler) SubmitResultHandler(w http.ResponseWriter, r *http.Request) {
	var result models.TaskResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.taskService.SubmitResult(result.ID, result.Result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to submit result"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{Message: "result accepted"})
}

// Вспомогательная функция для получения времени операции
func getOperationTime(operation string) int {
	// Используем переменные окружения из конфигурации
	switch operation {
	case "+":
		return config.Get().AddTime.Milliseconds()
	case "-":
		return config.Get().SubTime.Milliseconds()
	case "*":
		return config.Get().MulTime.Milliseconds()
	case "/":
		return config.Get().DivTime.Milliseconds()
	default:
		return 0
	}
}
