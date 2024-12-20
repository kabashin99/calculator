package handler

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalculateHandler(t *testing.T) {
	log.Println("Запуск тестов CalculateHandler")
	reqBody := `{"expression": "3 + 5"}`
	req, err := http.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	CalculateHandler(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %v", res.StatusCode)
	}
	log.Println("Тесты CalculateHandler выполнены")
}
