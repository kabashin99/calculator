package handler

import (
	"bytes"
	"calculator_app/internal/orchestrator/repository"
	"calculator_app/internal/pkg/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

type MockOrchestrator struct{}

func (m *MockOrchestrator) RegisterUser(user models.User) error {
	if user.Login == "existingUser" {
		return fmt.Errorf("user already exists")
	}
	return nil
}

func (m *MockOrchestrator) Authenticate(login, password string) (string, time.Time, error) {
	if login == "validUser" && password == "validPassword" {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"login": login,
		})
		tokenString, _ := token.SignedString([]byte("secret"))
		exp := time.Now().Add(time.Hour * 1)
		return tokenString, exp, nil
	}
	return "", time.Time{}, fmt.Errorf("invalid credentials")
}

func (m *MockOrchestrator) UserExists(login string) (bool, error) {
	return login == "validUser", nil
}

func (m *MockOrchestrator) AddExpression(expression, owner string) (string, error) {
	if owner == "validUser" {
		return "123", nil
	}
	return "", fmt.Errorf("error adding expression")
}

func (m *MockOrchestrator) GetExpressions(owner string) (map[string]*models.Expression, error) {
	return map[string]*models.Expression{
		"123": {
			ID:     "123",
			Owner:  owner,
			Status: repository.TaskStatusPending,
			Result: nil,
		},
	}, nil
}

func (m *MockOrchestrator) GetExpressionByID(id, owner string) (*models.Expression, bool, error) {
	if id == "123" && owner == "validUser" {
		return &models.Expression{
			ID:     "123",
			Owner:  owner,
			Status: "calculated",
			Result: nil,
		}, true, nil
	}

	return nil, false, fmt.Errorf("expression not found")
}

func TestRegisterUser(t *testing.T) {
	orc := &MockOrchestrator{}
	handler := NewHandler(orc)

	user := models.User{Login: "newUser", Password: "password123"}
	userJson, _ := json.Marshal(user)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(userJson))
	w := httptest.NewRecorder()

	handler.RegisterUser(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, response["status"], "created successfully")
}

func TestLoginUser(t *testing.T) {
	orc := &MockOrchestrator{}
	handler := NewHandler(orc)

	creds := struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}{
		Login:    "validUser",
		Password: "validPassword",
	}
	credsJson, _ := json.Marshal(creds)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(credsJson))
	w := httptest.NewRecorder()

	handler.LoginUser(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, response["token"].(string), "ey")
	assert.Contains(t, response["expires_at"].(string), "2025-05-11T")
}

func TestGetExpressions(t *testing.T) {
	orc := &MockOrchestrator{}
	handler := NewHandler(orc)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"login": "validUser"})
	tokenString, _ := token.SignedString([]byte(""))

	req := httptest.NewRequest("GET", "/expressions", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	handler.GetExpressions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Expressions []map[string]interface{} `json:"expressions"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Len(t, response.Expressions, 1)

	expr := response.Expressions[0]
	assert.Equal(t, "123", expr["id"])
	assert.Equal(t, "pending", expr["status"])
	assert.Equal(t, "validUser", expr["owner"])
	assert.Nil(t, expr["result"])
}
