package handler

import (
	"bytes"
	"calculator_app/internal/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCalculateHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		reqBody := []byte(`{"expression": "3 + 5"}`)
		req, err := http.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		CalculateHandler(w, req)

		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %v", res.StatusCode)
		}

		var response models.SuccessResponse
		err = json.NewDecoder(res.Body).Decode(&response)
		if err != nil {
			t.Errorf("error decoding response: %v", err)
		}
		if response.Result != "8.000000" {
			t.Errorf("expected result 8.000000, got %s", response.Result)
		}

	})

	t.Run("405 Method Not Allowed", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/calculate", nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		CalculateHandler(w, req)

		res := w.Result()
		if res.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %v", res.StatusCode)
		}

		var response models.ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&response)
		if err != nil {
			t.Errorf("error decoding response: %v", err)
		}
		if response.Error != "метод не разрешен" {
			t.Errorf("expected error message 'метод не разрешен', got '%s'", response.Error)
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBufferString("invalid json"))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		CalculateHandler(w, req)

		res := w.Result()
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %v", res.StatusCode)
		}

		var response models.ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&response)
		if err != nil {
			t.Errorf("error decoding response: %v", err)
		}
		if !strings.Contains(response.Error, "ошибка декодирования JSON") {
			t.Errorf("expected error message containing 'ошибка декодирования JSON', got '%s'", response.Error)
		}
	})

	t.Run("422 Unprocessable Entity - Invalid Characters", func(t *testing.T) {
		reqBody := []byte(`{"expression": "а +* 5"}`)
		req, err := http.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		CalculateHandler(w, req)

		res := w.Result()
		if res.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("expected status 422, got %v", res.StatusCode)
		}

		var response models.ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&response)
		if err != nil {
			t.Errorf("error decoding response: %v", err)
		}
		if response.Error != "недопустимые символы" {
			t.Errorf("expected error message 'недопустимые символы', got '%s'", response.Error)
		}
	})

	t.Run("422 Unprocessable Entity - Calculation Error", func(t *testing.T) {
		reqBody := []byte(`{"expression": "10 / 0"}`) // пример выражения, вызывающего ошибку
		req, err := http.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		CalculateHandler(w, req)

		res := w.Result()
		if res.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("expected status 422, got %v", res.StatusCode)
		}

		var response models.ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&response)
		if err != nil {
			t.Errorf("error decoding response: %v", err)
		}
		if !strings.Contains(response.Error, "деление на ноль") { // сообщение об ошибке может меняться в зависимости от реализации Calc
			t.Errorf("expected error message containing 'деление на ноль', got '%s'", response.Error)
		}
	})
}
