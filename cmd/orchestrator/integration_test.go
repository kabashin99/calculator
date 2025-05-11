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

func startServers(t *testing.T) (httpURL, grpcAddr string, cleanup func()) {
	dbConn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RunMigrations(dbConn); err != nil {
		t.Fatal(err)
	}

	repo := repository.NewRepository(dbConn)
	orcSvc := service.NewOrchestrator(1, 1, 1, 1, repo)
	h := handler.NewHandler(orcSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/register", h.RegisterUser)
	mux.HandleFunc("/api/v1/login", h.LoginUser)
	mux.HandleFunc("/api/v1/calculate", h.AddExpression)
	mux.HandleFunc("/api/v1/expressions", h.GetExpressions)
	mux.HandleFunc("/api/v1/expressions/", h.GetExpressionByID)
	httpSrv := httptest.NewServer(mux)

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

	creds := map[string]string{"login": "alice", "password": "pass"}
	b, _ := json.Marshal(creds)
	if resp, err := http.Post(httpURL+"/api/v1/register", "application/json", bytes.NewReader(b)); err != nil {
		t.Fatal(err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("register failed: %v", resp.Status)
	}

	if resp, err := http.Post(httpURL+"/api/v1/login", "application/json", bytes.NewReader(b)); err != nil {
		t.Fatal(err)
	} else {
		var lr struct {
			Token string `json:"token"`
		}
		json.NewDecoder(resp.Body).Decode(&lr)
		if lr.Token == "" {
			t.Fatal("empty token")
		}
		creds["token"] = lr.Token
	}

	exprReq := map[string]string{"expression": "2+3"}
	b, _ = json.Marshal(exprReq)
	req, _ := http.NewRequest("POST", httpURL+"/api/v1/calculate", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+creds["token"])
	if resp, err := http.DefaultClient.Do(req); err != nil {
		t.Fatal(err)
	} else if resp.StatusCode != http.StatusCreated {
		t.Fatalf("calculate failed: %v", resp.Status)
	} else {
		var cr struct {
			ID string `json:"id"`
		}
		json.NewDecoder(resp.Body).Decode(&cr)
		conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		cli := pb.NewOrchestratorServiceClient(conn)

		getResp, err := cli.GetTask(context.Background(), &pb.GetTaskRequest{})
		if err != nil {
			t.Fatal(err)
		}

		_, err = cli.SubmitResult(context.Background(), &pb.SubmitResultRequest{
			TaskId:  getResp.TaskId,
			Outcome: &pb.SubmitResultRequest_Result{Result: getResp.Arg1 + getResp.Arg2},
		})
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(100 * time.Millisecond)

		resResp, err := cli.GetTaskResult(context.Background(), &pb.GetTaskResultRequest{TaskId: getResp.TaskId})
		if err != nil {
			t.Fatal(err)
		}
		if !resResp.TaskExists {
			t.Fatal("task result does not exist")
		}
		if resResp.Result.GetValue() != 5 {
			t.Fatalf("expected gRPC result=5, got %v", resResp.Result.GetValue())
		}
	}
}
