//go:build integration
// +build integration

package main

import (
	"bytes"
	"calculator_app/db"
	orchestratorgrpc "calculator_app/internal/orchestrator/grpc"
	"calculator_app/internal/orchestrator/handler"
	"calculator_app/internal/orchestrator/repository"
	"calculator_app/internal/orchestrator/service"
	pb "calculator_app/internal/proto"
	"context"
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"google.golang.org/grpc"
	_ "modernc.org/sqlite"
)

// helper: spin up HTTP + gRPC servers on random ports
func startServers(t *testing.T) (httpURL string, grpcAddr string, cleanup func()) {
	// в памяти
	dbConn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RunMigrations(dbConn); err != nil {
		t.Fatal(err)
	}

	repo := repository.NewRepository(dbConn)
	// тайминги операций не важны
	orcSvc := service.NewOrchestrator(1, 1, 1, 1, repo)
	h := handler.NewHandler(orcSvc)

	// HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/register", h.RegisterUser)
	mux.HandleFunc("/api/v1/login", h.LoginUser)
	mux.HandleFunc("/api/v1/calculate", h.AddExpression)
	mux.HandleFunc("/api/v1/expressions", h.GetExpressions)
	mux.HandleFunc("/api/v1/expressions/", h.GetExpressionByID)

	httpSrv := httptest.NewServer(mux)

	// gRPC
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterOrchestratorServiceServer(grpcSrv, orchestratorgrpc.NewOrchestratorGRPCServer(orcSvc))
	go grpcSrv.Serve(lis)

	return httpSrv.URL, lis.Addr().String(), func() {
		httpSrv.Close()
		grpcSrv.Stop()
		dbConn.Close()
	}
}

func TestEndToEnd(t *testing.T) {
	httpURL, grpcAddr, cleanup := startServers(t)
	defer cleanup()

	// 1) Регистрация
	regBody := map[string]string{"login": "alice", "password": "pass"}
	b, _ := json.Marshal(regBody)
	resp, err := http.Post(httpURL+"/api/v1/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("register failed: %v", resp.Status)
	}

	// 2) Логин и получение токена
	b, _ = json.Marshal(regBody)
	resp, err = http.Post(httpURL+"/api/v1/login", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	var loginResp struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expires_at"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	if loginResp.Token == "" {
		t.Fatal("empty token")
	}

	// 3) Добавить выражение «2+3»
	reqBody := map[string]string{"expression": "2+3"}
	b, _ = json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", httpURL+"/api/v1/calculate", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("calculate failed: %v", resp.Status)
	}
	var calcResp struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&calcResp)

	// 4) Агент: забирает задачу по gRPC
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	cli := pb.NewOrchestratorServiceClient(conn)

	// Получаем задачу
	getResp, err := cli.GetTask(context.Background(), &pb.GetTaskRequest{})
	if err != nil {
		t.Fatal(err)
	}

	// 5) Агент: отправляет результат (2+3=5)
	_, err = cli.SubmitResult(context.Background(), &pb.SubmitResultRequest{
		TaskId:  getResp.TaskId,
		Outcome: &pb.SubmitResultRequest_Result{Result: getResp.Arg1 + getResp.Arg2},
	})
	if err != nil {
		t.Fatal(err)
	}

	// даём время на запись финального результата
	time.Sleep(50 * time.Millisecond)

	// 6) Проверяем финальный результат через HTTP
	req, _ = http.NewRequest("GET", httpURL+"/api/v1/expressions/"+calcResp.ID, nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var exprResp struct {
		Expression struct {
			ID     string   `json:"id"`
			Status string   `json:"status"`
			Result *float64 `json:"result"`
		} `json:"expression"`
	}
	json.NewDecoder(resp.Body).Decode(&exprResp)
	if exprResp.Expression.Result == nil || *exprResp.Expression.Result != 5 {
		t.Fatalf("expected result=5, got %v", exprResp.Expression.Result)
	}
}
