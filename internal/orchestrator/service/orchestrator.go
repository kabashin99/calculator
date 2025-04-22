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
	expressions    map[string]*models.Expression
	tasks          []*models.Task
	completedTasks map[string]*models.Task
	mu             sync.Mutex

	timeAdditionMS       int
	timeSubtractionMS    int
	timeMultiplicationMS int
	timeDivisionMS       int
}

func NewOrchestrator(timeAdditionMS, timeSubtractionMS, timeMultiplicationMS, timeDivisionMS int) *Orchestrator {
	return &Orchestrator{
		expressions:          make(map[string]*models.Expression),
		tasks:                make([]*models.Task, 0),
		completedTasks:       make(map[string]*models.Task),
		timeAdditionMS:       timeAdditionMS,
		timeSubtractionMS:    timeSubtractionMS,
		timeMultiplicationMS: timeMultiplicationMS,
		timeDivisionMS:       timeDivisionMS,
	}
}

func (o *Orchestrator) AddExpression(userID string, expression string) (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	id := generateUUID()
	o.expressions[id] = &models.Expression{
		ID:     id,
		UserID: userID,
		Status: "pending",
		Result: 0,
	}

	log.Printf("New expression added: ID=%s, UserID=%s, Expression=%s", id, userID, expression)

	tasks, err := o.parseExpressionToTasks(expression, id)
	if err != nil {
		return "", err
	}

	o.tasks = append(o.tasks, tasks...)
	log.Printf("Tasks created for expression ID=%s: %+v", id, tasks)

	log.Printf("All tasks after adding expression: %+v", o.tasks)

	return id, nil
}

func (o *Orchestrator) parseExpressionToTasks(expression, expressionID string) ([]*models.Task, error) {
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

	log.Println("GetTask called. Checking available tasks...")

	if len(o.tasks) == 0 {
		log.Println("No tasks available")
		return nil, false
	}

	task := o.tasks[0]
	o.tasks = o.tasks[1:]

	log.Printf("Task to be dispatched: %+v", task)

	taskCopy := *task
	if o.completedTasks == nil {
		o.completedTasks = make(map[string]*models.Task)
	}
	o.completedTasks[taskCopy.ID] = &taskCopy

	log.Printf("Task dispatched: %+v", taskCopy)
	return &taskCopy, true
}

func (o *Orchestrator) SubmitResult(taskID string, result float64) bool {
	o.mu.Lock()
	defer o.mu.Unlock()

	log.Printf("Received result submission: Task ID=%s, Result=%f", taskID, result)

	log.Printf("Current expressions: %+v", o.expressions)
	log.Printf("Current completed tasks: %+v", o.completedTasks)

	parts := strings.Split(taskID, "-")
	if len(parts) < 5 {
		log.Printf("Invalid task ID: %s", taskID)
		return false
	}
	expressionID := strings.Join(parts[:5], "-")

	log.Printf("Extracted expression ID: %s from task ID: %s", expressionID, taskID)

	expr, exists := o.expressions[expressionID]
	if !exists {
		log.Printf("Expression not found for task ID: %s", taskID)
		return false
	}

	if task, ok := o.completedTasks[taskID]; ok {
		task.Result = result
		log.Printf("Task updated: ID=%s, Result=%f", taskID, result)
	} else {
		log.Printf("Task not found in completedTasks: %s", taskID)
		return false
	}

	allTasksDone := true
	var finalResult float64

	for id, task := range o.completedTasks {
		if strings.HasPrefix(id, expressionID) {
			if task.Result == 0 {
				allTasksDone = false
				break
			}

			isFinalTask := true
			for _, t := range o.completedTasks {
				for _, dep := range t.DependsOn {
					if dep == id {
						isFinalTask = false
						break
					}
				}
				if !isFinalTask {
					break
				}
			}

			if isFinalTask {
				finalResult = task.Result
			}
		}
	}

	if allTasksDone {
		expr.Result = finalResult
		expr.Status = "done"
		log.Printf("Expression completed: ID=%s, Result=%f, Status=%s",
			expressionID, finalResult, expr.Status)
	}

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
