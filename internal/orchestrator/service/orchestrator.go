package service

import (
	"calculator_app/internal/pkg/models"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Orchestrator struct {
	expressions map[string]*models.Expression
	tasks       []*models.Task
	mu          sync.Mutex

	timeAdditionMS       int
	timeSubtractionMS    int
	timeMultiplicationMS int
	timeDivisionMS       int
}

func NewOrchestrator(timeAdditionMS, timeSubtractionMS, timeMultiplicationMS, timeDivisionMS int) *Orchestrator {
	return &Orchestrator{
		expressions:          make(map[string]*models.Expression),
		tasks:                make([]*models.Task, 0),
		timeAdditionMS:       timeAdditionMS,
		timeSubtractionMS:    timeSubtractionMS,
		timeMultiplicationMS: timeMultiplicationMS,
		timeDivisionMS:       timeDivisionMS,
	}
}

func (o *Orchestrator) AddExpression(expression string) (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	id := generateUUID()
	o.expressions[id] = &models.Expression{
		ID:     id,
		Status: "pending",
		Result: 0,
	}

	log.Printf("New expression added: ID=%s, Expression=%s", id, expression)

	// Разбиваем выражение на задачи
	tasks, err := o.parseExpressionToTasks(expression, id)
	if err != nil {
		return "", err
	}

	o.tasks = append(o.tasks, tasks...)
	log.Printf("Tasks created for expression ID=%s: %+v", id, tasks)

	return id, nil
}

// internal/orchestrator/service/orchestrator.go

func (o *Orchestrator) parseExpressionToTasks(expression, expressionID string) ([]*models.Task, error) {
	postfix, err := shuntingYard(expression)
	if err != nil {
		return nil, fmt.Errorf("shunting yard error: %v", err)
	}

	var tasks []*models.Task
	taskMap := make(map[string]*models.Task)
	var stack []string

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
			}

			// Обработка зависимостей
			if strings.HasPrefix(left, "task:") {
				depID := strings.TrimPrefix(left, "task:")
				task.DependsOn = append(task.DependsOn, depID)
				left = "0" // Временное значение
			}
			if strings.HasPrefix(right, "task:") {
				depID := strings.TrimPrefix(right, "task:")
				task.DependsOn = append(task.DependsOn, depID)
				right = "0" // Временное значение
			}

			task.Arg1 = parseFloat(left)
			task.Arg2 = parseFloat(right)
			task.OperationTime = o.getOperationTime(token)

			tasks = append(tasks, task)
			taskMap[taskID] = task
			stack = append(stack, "task:"+taskID)
		}
	}

	// Упорядочиваем задачи по зависимостям
	orderedTasks := topologicalSort(tasks, taskMap)
	return orderedTasks, nil
}

func topologicalSort(tasks []*models.Task, taskMap map[string]*models.Task) []*models.Task {
	var result []*models.Task
	visited := make(map[string]bool)

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

	return result
}

// parseFloat преобразует строку в float64
func parseFloat(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

// shuntingYard преобразует выражение в обратную польскую запись
func shuntingYard(expression string) ([]string, error) {
	var output []string
	var operators []string

	tokens := tokenize(expression) // Разбиваем выражение на токены

	for _, token := range tokens {
		if isNumber(token) {
			output = append(output, token)
		} else if isOperator(token) {
			for len(operators) > 0 && precedence(operators[len(operators)-1]) >= precedence(token) {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			operators = append(operators, token)
		} else if token == "(" {
			operators = append(operators, token)
		} else if token == ")" {
			for len(operators) > 0 && operators[len(operators)-1] != "(" {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			if len(operators) == 0 || operators[len(operators)-1] != "(" {
				return nil, fmt.Errorf("mismatched parentheses")
			}
			operators = operators[:len(operators)-1]
		} else {
			return nil, fmt.Errorf("invalid token: %s", token)
		}
	}

	for len(operators) > 0 {
		if operators[len(operators)-1] == "(" || operators[len(operators)-1] == ")" {
			return nil, fmt.Errorf("mismatched parentheses")
		}
		output = append(output, operators[len(operators)-1])
		operators = operators[:len(operators)-1]
	}

	return output, nil
}

// precedence возвращает приоритет оператора
func precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}

// isNumber проверяет, является ли токен числом
func isNumber(token string) bool {
	_, err := strconv.ParseFloat(token, 64)
	return err == nil
}

// isOperator проверяет, является ли токен оператором
func isOperator(token string) bool {
	return token == "+" || token == "-" || token == "*" || token == "/"
}

func (o *Orchestrator) GetExpressions() map[string]*models.Expression {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.expressions
}

func (o *Orchestrator) GetExpressionByID(id string) (*models.Expression, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	expr, exists := o.expressions[id]
	return expr, exists
}

func (o *Orchestrator) GetTask() (*models.Task, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if len(o.tasks) == 0 {
		log.Println("No tasks available")
		return nil, false
	}

	task := o.tasks[0]
	o.tasks = o.tasks[1:]
	log.Printf("Task dispatched: %+v", task)
	return task, true
}

func (o *Orchestrator) SubmitResult(taskID string, result float64) bool {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Находим выражение по ID задачи
	expressionID := strings.Split(taskID, "-")[0] // Предполагаем, что ID задачи имеет формат "expressionID-taskNumber"
	expr, exists := o.expressions[expressionID]
	if !exists {
		log.Printf("Expression not found for task ID: %s", taskID)
		return false
	}

	// Обновляем результат задачи
	for i, task := range o.tasks {
		if task.ID == taskID {
			o.tasks[i].Result = result
			break
		}
	}

	// Обновляем результат выражения
	expr.Result = result
	expr.Status = "done"
	log.Printf("Result updated for expression ID=%s: Result=%f, Status=%s", expressionID, result, expr.Status)
	return true
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

// tokenize разбивает выражение на токены
func tokenize(expression string) []string {
	var tokens []string
	var currentToken strings.Builder

	for _, char := range expression {
		if char == ' ' {
			continue // Пропускаем пробелы
		}

		if isOperator(string(char)) || char == '(' || char == ')' {
			// Если текущий токен не пуст, добавляем его в список токенов
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			// Добавляем оператор или скобку как отдельный токен
			tokens = append(tokens, string(char))
		} else {
			// Добавляем символ к текущему токену (число)
			currentToken.WriteRune(char)
		}
	}

	// Добавляем последний токен, если он есть
	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return tokens
}
