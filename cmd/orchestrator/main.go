package main

import (
	"calculator_app/internal/orchestrator/handler"
	"calculator_app/internal/orchestrator/repository"
	"calculator_app/internal/orchestrator/service"
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

func main() {
	// Открываем базу данных
	db, err := sql.Open("sqlite", "./database.db?_journal=WAL&_busy_timeout=5000")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Инициализируем репозитории
	exprRepo := repository.NewExpressionRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// Инициализируем сервисы
	exprService := service.NewExpressionService(exprRepo, taskRepo)
	taskService := service.NewTaskService(taskRepo, exprRepo)

	// Создаём маршрутизатор
	router := mux.NewRouter()

	// Регистрируем обработчики
	handler.InitExpressionHandlers(router, exprService)
	handler.InitTaskHandlers(router, taskService)

	// Запускаем сервер
	log.Println("Оркестратор запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
