package main

import (
	"calculator_app/db"
	"calculator_app/internal/config"
	grpcservice "calculator_app/internal/orchestrator/grpc"
	"calculator_app/internal/orchestrator/handler"
	"calculator_app/internal/orchestrator/repository"
	"calculator_app/internal/orchestrator/service"
	pb "calculator_app/internal/proto"
	"database/sql"
	"google.golang.org/grpc"
	"log"
	_ "modernc.org/sqlite"
	"net"
	"net/http"
	"os"
)

func main() {
	cfg, err := config.LoadConfig("config/config.txt")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := os.MkdirAll("db", os.ModePerm); err != nil {
		log.Fatalf("Failed to create db directory: %v", err)
	}

	connStr := "file:db/calculator.db?cache=shared&mode=rwc"
	dbConn, err := sql.Open("sqlite", connStr)
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}

	defer func() {
		if err := dbConn.Close(); err != nil {
			log.Printf("Warning: failed to close DB: %v", err)
		}
	}()

	_, err = dbConn.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Fatalf("Failed to set WAL mode: %v", err)
	}

	_, err = dbConn.Exec("PRAGMA busy_timeout = 5000;") // 5 секунд
	if err != nil {
		log.Fatalf("Failed to set busy_timeout: %v", err)
	}

	if err := dbConn.Ping(); err != nil {
		log.Fatalf("DB Ping failed: %v", err)
	}

	if err := db.RunMigrations(dbConn); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	repo := repository.NewRepository(dbConn)
	orc := service.NewOrchestrator(cfg.TimeAdditionMS, cfg.TimeSubtractionMS, cfg.TimeMultiplicationMS, cfg.TimeDivisionMS, repo)
	OrchHandler := handler.NewHandler(orc)

	http.HandleFunc("POST /api/v1/register", OrchHandler.RegisterUser)
	http.HandleFunc("POST /api/v1/login", OrchHandler.LoginUser)
	http.HandleFunc("POST /api/v1/calculate", OrchHandler.AddExpression)
	http.HandleFunc("GET /api/v1/expressions", OrchHandler.GetExpressions)
	http.HandleFunc("GET /api/v1/expressions/{id}", OrchHandler.GetExpressionByID)

	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		pb.RegisterOrchestratorServiceServer(s, grpcservice.NewOrchestratorGRPCServer(orc))

		log.Println("gRPC сервер запущен на порту 50051")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	log.Println("HTTP сервер запущен на localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
