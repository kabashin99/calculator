package service

import (
	"calculator_app/internal/config"
	"calculator_app/internal/orchestrator/repository"
	"calculator_app/internal/pkg/models"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strconv"
	"strings"
	"time"
)

type Orchestrator struct {
	//repo                 *repository.Repository
	repo                 repository.RepositoryInterface
	timeAdditionMS       int
	timeSubtractionMS    int
	timeMultiplicationMS int
	timeDivisionMS       int
}

type OrchestratorInterface interface {
	RegisterUser(user models.User) error
	Authenticate(login, password string) (string, time.Time, error)
	UserExists(login string) (bool, error)
	AddExpression(expr string, login string) (string, error)
	GetExpressions(owner string) (map[string]*models.Expression, error)
	GetExpressionByID(id, owner string) (*models.Expression, bool, error)
}

func NewOrchestrator(timeAdditionMS, timeSubtractionMS, timeMultiplicationMS, timeDivisionMS int, repo repository.RepositoryInterface) *Orchestrator {
	return &Orchestrator{
		repo:                 repo,
		timeAdditionMS:       timeAdditionMS,
		timeSubtractionMS:    timeSubtractionMS,
		timeMultiplicationMS: timeMultiplicationMS,
		timeDivisionMS:       timeDivisionMS,
	}
}

func (o *Orchestrator) AddExpression(expression string, owner string) (string, error) {
	id := generateUUID()
	err := o.repo.AddExpression(&models.Expression{
		ID:     id,
		Status: repository.TaskStatusPending,
		Result: nil,
		Owner:  owner,
	})

	if err != nil {
		return "", fmt.Errorf("failed to save expression: %w", err)
	}

	tasks, err := o.parseExpressionToTasks(expression, id, owner)
	if err != nil {
		return "", err
	}

	for _, task := range tasks {
		task.UserLogin = owner
		if err := o.repo.AddTask(task); err != nil {
			return "", fmt.Errorf("failed to add task: %w", err)
		}
	}

	return id, nil
}

func (o *Orchestrator) parseExpressionToTasks(
	expression string,
	expressionID string,
	owner string,

) ([]*models.Task, error) {

	postfix, err := shuntingYard(tokenize(expression))
	if err != nil {
		return nil, fmt.Errorf("shunting yard error: %v", err)
	}

	var tasks []*models.Task
	taskMap := make(map[string]*models.Task)
	var stack []string

	log.Printf("Parsing expression: %s", expression)
	log.Printf("Postfix notation: %v", postfix)

	for _, token := range postfix {
		if isNumber(token) {
			stack = append(stack, token)
			continue
		}

		if isOperator(token) {
			if len(stack) < 2 {
				return nil, fmt.Errorf("not enough operands for operator %s", token)
			}

			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			taskID := fmt.Sprintf("%s-%d", expressionID, len(tasks)+1)
			task := &models.Task{
				ID:        taskID,
				Operation: token,
				DependsOn: []string{},
				UserLogin: owner,
			}

			if strings.HasPrefix(left, "task:") {
				depID := strings.TrimPrefix(left, "task:")
				task.DependsOn = append(task.DependsOn, depID)
				left = "0"
			}
			if strings.HasPrefix(right, "task:") {
				depID := strings.TrimPrefix(right, "task:")
				task.DependsOn = append(task.DependsOn, depID)
				right = "0"
			}

			task.Arg1 = parseFloat(left)
			task.Arg2 = parseFloat(right)
			task.OperationTime = o.getOperationTime(token)

			tasks = append(tasks, task)
			taskMap[taskID] = task
			stack = append(stack, "task:"+taskID)

			log.Printf("Created task: %+v", task)
		}
	}

	orderedTasks := topologicalSort(tasks, taskMap)
	log.Printf("Ordered tasks: %+v", orderedTasks)

	return orderedTasks, nil
}

func topologicalSort(tasks []*models.Task, taskMap map[string]*models.Task) []*models.Task {
	visited := make(map[string]bool)
	result := make([]*models.Task, 0, len(tasks))

	var visit func(string)
	visit = func(taskID string) {
		if visited[taskID] {
			return
		}
		visited[taskID] = true
		for _, dep := range taskMap[taskID].DependsOn {
			visit(dep)
		}
		result = append(result, taskMap[taskID])
	}

	for _, task := range tasks {
		if !visited[task.ID] {
			visit(task.ID)
		}
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

func parseFloat(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

func shuntingYard(tokens []string) ([]string, error) {
	var output []string
	var operators []string

	precedence := map[string]int{
		"+": 1, "-": 1,
		"*": 2, "/": 2,
	}

	for _, token := range tokens {
		if isNumber(token) {
			output = append(output, token)
		} else if isOperator(token) {
			for len(operators) > 0 {
				top := operators[len(operators)-1]
				if precedence[top] >= precedence[token] {
					output = append(output, top)
					operators = operators[:len(operators)-1]
				} else {
					break
				}
			}
			operators = append(operators, token)
		} else if token == "(" {
			operators = append(operators, token)
		} else if token == ")" {
			for len(operators) > 0 {
				top := operators[len(operators)-1]
				operators = operators[:len(operators)-1]
				if top == "(" {
					break
				}
				output = append(output, top)
			}
		} else {
			return nil, fmt.Errorf("invalid token: %s", token)
		}
	}

	for len(operators) > 0 {
		top := operators[len(operators)-1]
		operators = operators[:len(operators)-1]
		if top == "(" || top == ")" {
			return nil, fmt.Errorf("mismatched parentheses")
		}
		output = append(output, top)
	}

	return output, nil
}

func isNumber(token string) bool {
	_, err := strconv.ParseFloat(token, 64)
	return err == nil
}

func isOperator(token string) bool {
	return token == "+" || token == "-" || token == "*" || token == "/"
}

func (o *Orchestrator) GetExpressions(owner string) (map[string]*models.Expression, error) {
	return o.repo.GetExpressionsByOwner(owner)
}

func (o *Orchestrator) GetExpressionByID(id string, owner string) (*models.Expression, bool, error) {
	return o.repo.GetExpressionByIDAndOwner(id, owner)
}

func (o *Orchestrator) GetTask() (*models.Task, bool, error) {
	task, exists, err := o.repo.GetAndLockTask()
	if err != nil {
		log.Printf("Repository error: %v", err)
		return nil, false, err
	}

	if !exists {
	} else {
		log.Printf("Found task: %+v", task)
	}

	return task, exists, nil
}

func (o *Orchestrator) SubmitResult(taskID string, result float64, taskErr *models.TaskError) (bool, error) {
	var resultPtr *float64
	if taskErr == nil {
		resultPtr = &result
	}

	updated, status, err := o.repo.UpdateTaskResult(taskID, resultPtr, taskErr)
	if err != nil {
		return false, fmt.Errorf("failed to update task: %w", err)
	}
	if !updated {
		return false, nil
	}

	parts := strings.Split(taskID, "-")
	if len(parts) < 6 {
		return false, fmt.Errorf("invalid taskID format: %s", taskID)
	}
	exprID := strings.Join(parts[:5], "-")

	if status != repository.TaskStatusCompleted {
		_, _ = o.repo.UpdateExpression(exprID, status, 0)
		return true, nil
	}

	allDone, err := o.repo.AreAllTasksCompleted(exprID)
	if err != nil {
		return false, fmt.Errorf("failed to check tasks: %w", err)
	}

	if !allDone {
		return true, nil
	}

	finalResult, err := o.repo.CalculateFinalResult(exprID)
	if err != nil {
		return false, fmt.Errorf("failed to calculate result: %w", err)
	}

	exprUpdated, err := o.repo.UpdateExpression(exprID, repository.TaskStatusCompleted, finalResult)
	if err != nil {
		return false, fmt.Errorf("failed to update expression: %w", err)
	}
	if !exprUpdated {
		log.Printf("Failed to update expression with ID %s", exprID)
		return false, fmt.Errorf("expression not found or not updated")
	}

	return true, nil
}

func generateUUID() string {
	return uuid.New().String()
}

func (o *Orchestrator) getOperationTime(operation string) int {
	switch operation {
	case "+":
		return o.timeAdditionMS
	case "-":
		return o.timeSubtractionMS
	case "*":
		return o.timeMultiplicationMS
	case "/":
		return o.timeDivisionMS
	default:
		return 0
	}
}

func tokenize(expression string) []string {
	var tokens []string
	var currentToken strings.Builder

	for _, char := range expression {
		if char == ' ' {
			continue
		}

		if isOperator(string(char)) || char == '(' || char == ')' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(char))
		} else {
			currentToken.WriteRune(char)
		}
	}

	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return tokens
}

func (o *Orchestrator) RegisterUser(user models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)
	err = o.repo.RegisterUser(user)
	if err != nil {
		// Проверяем, содержит ли ошибка текст про уникальность логина
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.login") {
			return fmt.Errorf("user '%s' already exists", user.Login)
		}
		return fmt.Errorf("failed to register user: %w", err)
	}

	return nil
}

func (o *Orchestrator) Authenticate(login, password string) (string, time.Time, error) {
	user, err := o.repo.FindUser(login)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", time.Time{}, fmt.Errorf("invalid credentials")
	}

	tokenString, exp, err := generateToken(login)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, exp, nil
}

func (o *Orchestrator) UserExists(login string) (bool, error) {
	_, err := o.repo.FindUser(login)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (o *Orchestrator) GetTaskResult(taskID string) (float64, bool, error) {
	return o.repo.GetTaskResult(taskID)
}

func generateToken(userLogin string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(24 * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": userLogin,
		"exp":   exp.Unix(),
		"iat":   now.Unix(),
	})

	cfg, err := config.LoadConfig("config/config.txt")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	tokenString, err := token.SignedString([]byte(cfg.JwtSecretKey))
	if err != nil {
		log.Printf("Failed to generate token")
		return "", time.Time{}, fmt.Errorf("failed to generate token: %w", err)
	}
	return tokenString, exp, nil
}
