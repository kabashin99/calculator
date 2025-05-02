package handler

import (
	"calculator_app/internal/orchestrator/service"
	"calculator_app/internal/pkg/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const hmacSampleSecret = "super_secret_signature"

type Handler struct {
	orc   *service.Orchestrator
	users map[string]models.User
}

func NewHandler(orc *service.Orchestrator) *Handler {
	return &Handler{
		orc:   orc,
		users: make(map[string]models.User),
	}
}

func generateToken(userLogin string) string {
	now := time.Now()
	exp := now.Add(24 * time.Hour).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": userLogin, // Добавляем логин в claims
		"exp":   exp,
		"iat":   now.Unix(), // Время создания токена
	})

	tokenString, err := token.SignedString([]byte(hmacSampleSecret))
	if err != nil {
		log.Printf("Failed to generate token")
		return ""
	}
	return tokenString
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if _, exists := h.users[user.Login]; exists {
		http.Error(w, "User alredy exists", http.StatusConflict)
		return
	}

	h.users[user.Login] = user
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
	})
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	storedUSer, exists := h.users[user.Login]
	if !exists || storedUSer.Password != user.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := generateToken(user.Login)
	now := time.Now()
	exp := now.Add(24 * time.Hour).Unix()

	expTime := time.Unix(exp, 0).Format("2006-01-02 15:04:05")
	response := map[string]string{
		"token": "Bearer " + token,
		"exp":   expTime,
	}

	w.Header().Set("Content-Type", "application/json")
	//w.Header().Set("Authorization", "Bearer "+tokenString)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) CheckAuthorization(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is missing")
	}

	tokenString := authHeader[len("Bearer "):]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(hmacSampleSecret), nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token claims")
	}

	login, ok := claims["login"].(string)
	if !ok {
		return "", fmt.Errorf("missing login in token claims")
	}

	return login, nil
}

func (h *Handler) authorize(w http.ResponseWriter, r *http.Request) (string, error) {
	login, err := h.CheckAuthorization(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return "", err
	}
	return login, nil
}

func (h *Handler) AddExpression(w http.ResponseWriter, r *http.Request) {
	login, err := h.authorize(w, r)
	if err != nil {
		log.Printf("Error parsing token: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusUnprocessableEntity) // 422
		return
	}

	id, err := h.orc.AddExpression(req.Expression, login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity) //422
		return
	}

	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *Handler) GetExpressions(w http.ResponseWriter, r *http.Request) {

	login, err := h.authorize(w, r)
	if err != nil {
		return
	}

	exprMap := h.orc.GetExpressions()

	//expressions := make([]models.Expression, 0, len(exprMap))
	expressions := make([]models.Expression, 0)
	for _, expr := range exprMap {
		if expr.Owner == login {
			expressions = append(expressions, *expr)
		}
	}

	w.WriteHeader(http.StatusOK) //200
	json.NewEncoder(w).Encode(map[string]interface{}{
		"expressions": expressions,
	})
}

func (h *Handler) GetExpressionByID(w http.ResponseWriter, r *http.Request) {
	login, err := h.authorize(w, r)
	if err != nil {
		return
	}

	id := r.PathValue("id")
	expr, exists := h.orc.GetExpressionByID(id)
	if !exists {
		http.Error(w, "expression not found", http.StatusNotFound) //404
		return
	}

	if expr.Owner != login {
		http.Error(w, "Unauthorized access", http.StatusUnauthorized) // 401
		return
	}

	w.WriteHeader(http.StatusOK) // 200
	json.NewEncoder(w).Encode(map[string]interface{}{"expression": expr})
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	task, exists := h.orc.GetTask()
	if !exists {
		http.Error(w, "no tasks available", http.StatusNotFound) //404
		return
	}

	w.WriteHeader(http.StatusOK) // 200
	json.NewEncoder(w).Encode(map[string]interface{}{"task": task})
}

func (h *Handler) SubmitResult(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusUnprocessableEntity) // 422
		return
	}

	if !h.orc.SubmitResult(req.ID, req.Result) {
		http.Error(w, "task not found", http.StatusNotFound) // 404
		return
	}

	w.WriteHeader(http.StatusOK) // 200
}
