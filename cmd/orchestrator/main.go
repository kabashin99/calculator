package main

import (
	"calculator_app/internal/orchestrator/handler"
	"calculator_app/internal/orchestrator/repository"
	"calculator_app/internal/orchestrator/service"
	"log"
	"net/http"
)

func main() {
	exprRepo := repository.NewExpressionRepository()
	taskRepo := repository.NewTaskRepository()
	exprService := service.NewExpressionService(exprRepo, taskRepo)
	taskService := service.NewTaskService(taskRepo)

	handler.InitExpressionHandlers(exprService)
	handler.InitTaskHandlers(taskService)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
