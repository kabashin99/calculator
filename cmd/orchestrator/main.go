package main

import (
	"calculator_app/internal/orchestrator/handler"
	"calculator_app/internal/orchestrator/repository"
	"calculator_app/internal/orchestrator/service"
	"database/sql"
	"log"
	"net/http"
)

func main() {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	exprRepo := repository.NewExpressionRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	exprService := service.NewExpressionService(exprRepo, taskRepo)
	taskService := service.NewTaskService(taskRepo)

	handler.InitExpressionHandlers(exprService)
	handler.InitTaskHandlers(taskService)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
