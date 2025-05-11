package handler

import (
	"calculator_app/internal/config"
	"calculator_app/internal/orchestrator/service"
	"calculator_app/internal/pkg/models"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"time"
)

/*
type Handler struct {
	orc *service.Orchestrator
}

func NewHandler(orc *service.Orchestrator) *Handler {
	return &Handler{
		orc: orc,
	}
}
*/

type Handler struct {
	orc service.OrchestratorInterface
}

func NewHandler(orc service.OrchestratorInterface) *Handler {
	return &Handler{
		orc: orc,
	}
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.orc.RegisterUser(user); err != nil {
		http.Error(w, "Registration failed: "+err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
	userSuccessfully := fmt.Sprintf("user '%s' created successfully", user.Login)
	json.NewEncoder(w).Encode(map[string]string{"status": userSuccessfully})
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, exp, err := h.orc.Authenticate(creds.Login, creds.Password)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":      token,
		"expires_at": exp.Format(time.RFC3339),
	})
}

func (h *Handler) CheckAuthorization(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is missing")
	}

	cfg, err := config.LoadConfig("config/config.txt")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	tokenString := authHeader[len("Bearer "):]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(cfg.JwtSecretKey), nil
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
		//http.Error(w, "Unauthorized", http.StatusUnauthorized)
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

	exists, err := h.orc.UserExists(login)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "User not found", http.StatusForbidden)
		return
	}

	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}

	id, err := h.orc.AddExpression(req.Expression, login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *Handler) GetExpressions(w http.ResponseWriter, r *http.Request) {

	owner, err := h.authorize(w, r)
	if err != nil {
		return
	}

	exprMap, err := h.orc.GetExpressions(owner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	expressions := make([]models.Expression, 0)
	for _, expr := range exprMap {
		if expr.Owner == owner {
			expressions = append(expressions, *expr)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"expressions": expressions,
	})
}

func (h *Handler) GetExpressionByID(w http.ResponseWriter, r *http.Request) {
	owner, err := h.authorize(w, r)
	if err != nil {
		return
	}

	id := r.PathValue("id")
	expr, exists, err := h.orc.GetExpressionByID(id, owner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "expression not found", http.StatusNotFound)
		return
	}

	if expr.Owner != owner {
		http.Error(w, "Unauthorized access", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"expression": expr}); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
