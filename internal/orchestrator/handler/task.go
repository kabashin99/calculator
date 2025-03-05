package handler

import (
	"calculator_app/config"
	"calculator_app/internal/orchestrator/models"
	"calculator_app/internal/orchestrator/service"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var cfg *config.Config

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

func InitTaskHandlers(router *mux.Router, s *service.TaskService) {
	h := &TaskHandler{taskService: s}
	router.HandleFunc("/internal/task", h.GetNextTaskHandler).Methods("GET")
	router.HandleFunc("/internal/task", h.SubmitResultHandler).Methods("POST")
	router.HandleFunc("/tasks/{id}", h.GetTaskHandler).Methods("GET")
	router.HandleFunc("/tasks", h.GetTasksHandler).Methods("GET")

	cfg = config.Load()
}

func (h *TaskHandler) GetNextTaskHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🔍 Запрос на получение задачи...")

	task, err := h.taskService.GetNextTask()
	if err != nil {
		log.Println("❌ Ошибка при получении задачи:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
		return
	}

	if task == nil {
		log.Println("⚠️ Нет доступных задач")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "no tasks available"})
		return
	}

	// Проверяем, что аргументы задачи готовы
	if !h.taskService.IsTaskReady(task) {
		log.Println("⚠️ Задача не готова к выполнению:", task.ID)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "task dependencies not ready"})
		return
	}

	arg1, arg2, err := h.taskService.GetDependenciesResults(task)
	if err != nil {
		log.Printf("Error getting dependencies: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := models.TaskResponse{
		ID:            task.ID,
		Arg1:          task.Arg1,
		Arg2:          task.Arg2,
		Operation:     task.Operation,
		OperationTime: getOperationTime(task.Operation),
	}

	log.Printf("✅ Отдаю задачу: %v %s %v", response.Arg1, response.Operation, response.Arg2)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TaskHandler) SubmitResultHandler(w http.ResponseWriter, r *http.Request) {
	var result models.TaskResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		log.Println("❌ Ошибка парсинга JSON:", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid request body"})
		return
	}

	log.Printf("📥 Получен результат задачи %s: %f", result.ID, result.Result)

	if err := h.taskService.SubmitResult(result.ID, result.Result); err != nil {
		log.Println("❌ Ошибка при сохранении результата:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "failed to submit result"})
		return
	}

	log.Println("✅ Результат задачи успешно сохранён!")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{Message: "result accepted"})
}

func (h *TaskHandler) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	task, err := h.taskService.GetTask(taskID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "task not found"})
		return
	}

	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.taskService.GetAllTasks()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
		return
	}

	json.NewEncoder(w).Encode(tasks)
}

func getOperationTime(operation string) int {
	switch operation {
	case "+":
		return int(cfg.AddTime.Milliseconds())
	case "-":
		return int(cfg.SubTime.Milliseconds())
	case "*":
		return int(cfg.MulTime.Milliseconds())
	case "/":
		return int(cfg.DivTime.Milliseconds())
	default:
		return 0
	}
}
