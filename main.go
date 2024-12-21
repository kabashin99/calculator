// @title Calculator API
// @version 1.0
// @description HTTP-сервер, который обрабатывает входящие арифметические выражения и возвращает результаты вычислений
// @host localhost:8080
// @BasePath /api/v1

package main

import (
	_ "calculator_app/docs"
	"calculator_app/internal/handler"
	"calculator_app/internal/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
)

func main() {
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/v1/calculate", handler.CalculateHandler)

	swaggerMux := http.NewServeMux()
	swaggerMux.Handle("/swagger/", httpSwagger.WrapHandler)

	loggedMux := middleware.Logging(apiMux)

	go func() {
		log.Println("Сервер для API запущен на :8080")
		if err := http.ListenAndServe(":8080", loggedMux); err != nil {
			log.Fatalf("Ошибка при запуске сервера: %s\n", err)
		}
	}()

	log.Println("Сервер для Swagger запущен на :8081")
	if err := http.ListenAndServe(":8081", swaggerMux); err != nil {
		log.Fatalf("Ошибка при запуске сервера для Swagger: %s\n", err)
	}

	middleware.CloseLogFile()
}
