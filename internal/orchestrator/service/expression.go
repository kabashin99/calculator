package service

import (
	"calculator_app/internal/orchestrator/models"
	"calculator_app/internal/orchestrator/repository"
	"calculator_app/internal/pkg/calculator"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

type ExpressionService struct {
	exprRepo repository.ExpressionRepository
	taskRepo repository.TaskRepository
}

func NewExpressionService(exprRepo repository.ExpressionRepository, taskRepo repository.TaskRepository) *ExpressionService {
	return &ExpressionService{
		exprRepo: exprRepo,
		taskRepo: taskRepo,
	}
}

func ParseExpression(exprID, expression string) ([]*models.Task, error) {
	// Проверяем корректность скобок
	tokens, err := tokenize(expression)
	if err != nil {
		return nil, err
	}
	if !checkParenthesesBalance(tokens) {
		return nil, errors.New("некорректное количество скобок")
	}

	// Вызываем калькулятор, чтобы он разбил выражение на части
	steps, err := calculator.ParseToSteps(expression) // <-- Используем ваш calculator.go
	if err != nil {
		return nil, err
	}

	var tasks []*models.Task
	taskMap := make(map[string]string) // Храним ID задач по значениям

	for _, step := range steps {
		// Проверяем, есть ли аргументы в мапе (если уже были вычислены)
		arg1 := step.Arg1
		if id, found := taskMap[step.Arg1]; found {
			arg1 = id
		}

		arg2 := step.Arg2
		if id, found := taskMap[step.Arg2]; found {
			arg2 = id
		}

		// Создаем задачу
		task := &models.Task{
			ID:           uuid.New().String(),
			ExpressionID: exprID,
			Arg1:         arg1,
			Arg2:         arg2,
			Operation:    step.Operation,
			Status:       "pending",
		}
		tasks = append(tasks, task)

		// Запоминаем ID этой задачи
		taskMap[step.Result] = task.ID
	}

	return tasks, nil
}

// **1. Разбираем строку вручную на числа, операторы и скобки**
func tokenize(expression string) ([]string, error) {
	var tokens []string
	var numberBuffer strings.Builder

	for _, ch := range expression {
		switch {
		case ch >= '0' && ch <= '9' || ch == '.': // Число (целое или десятичное)
			numberBuffer.WriteRune(ch)

		case ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '(' || ch == ')':
			// Если перед оператором было число — добавляем его в токены
			if numberBuffer.Len() > 0 {
				tokens = append(tokens, numberBuffer.String())
				numberBuffer.Reset()
			}
			// Добавляем сам оператор или скобку
			tokens = append(tokens, string(ch))

		case ch == ' ': // Игнорируем пробелы
			if numberBuffer.Len() > 0 {
				tokens = append(tokens, numberBuffer.String())
				numberBuffer.Reset()
			}

		default:
			return nil, errors.New("invalid character in expression")
		}
	}

	// Добавляем последнее число (если есть)
	if numberBuffer.Len() > 0 {
		tokens = append(tokens, numberBuffer.String())
	}

	return tokens, nil
}

// **2. Проверяем корректность скобок**
func checkParenthesesBalance(tokens []string) bool {
	balance := 0
	for _, token := range tokens {
		if token == "(" {
			balance++
		} else if token == ")" {
			balance--
			if balance < 0 {
				return false // Закрывающая скобка раньше открывающей
			}
		}
	}
	return balance == 0
}

// **3. Преобразуем токены в последовательность задач**
func parseTokens(exprID string, tokens []string) ([]*models.Task, error) {
	var tasks []*models.Task
	var stack []interface{} // Числа, операторы, ID задач

	// Обрабатываем выражения в скобках сначала
	tokens, err := handleParentheses(exprID, tokens, &tasks)
	if err != nil {
		return nil, err
	}

	// **Сначала выполняем `* /` (высокий приоритет)**
	i := 0
	for i < len(tokens) {
		token := tokens[i]

		if token == "*" || token == "/" {
			if len(stack) < 1 {
				return nil, errors.New("некорректное выражение")
			}

			arg1 := stack[len(stack)-1]  // Берём последний элемент из стека
			stack = stack[:len(stack)-1] // Убираем его

			i++
			if i >= len(tokens) {
				return nil, errors.New("некорректное выражение")
			}

			arg2 := tokens[i] // Следующий токен — аргумент 2

			task := &models.Task{
				ID:           uuid.New().String(),
				ExpressionID: exprID,
				Arg1:         arg1,
				Arg2:         arg2,
				Operation:    token,
				Status:       "pending",
			}
			tasks = append(tasks, task)

			// Кладём ID новой задачи вместо аргумента
			stack = append(stack, task.ID)
		} else {
			stack = append(stack, token)
		}
		i++
	}

	// **Затем выполняем `+ -` (низкий приоритет)**
	for len(stack) > 1 {
		arg1 := stack[0]
		op := stack[1]
		arg2 := stack[2]
		stack = stack[3:]

		if opStr, ok := op.(string); ok && (opStr == "+" || opStr == "-") {
			task := &models.Task{
				ID:           uuid.New().String(),
				ExpressionID: exprID,
				Arg1:         arg1,
				Arg2:         arg2,
				Operation:    opStr,
				Status:       "pending",
			}
			tasks = append(tasks, task)

			// Кладём ID задачи в стек
			stack = append([]interface{}{task.ID}, stack...)
		} else {
			return nil, errors.New("неверный порядок операторов")
		}
	}

	return tasks, nil
}

// **6. Обрабатываем выражения в скобках**
func handleParentheses(exprID string, tokens []string, tasks *[]*models.Task) ([]string, error) {
	var result []string
	var stack []string

	for _, token := range tokens {
		if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			// Выполняем подвыражение
			subExpr := stack
			stack = []string{} // очищаем стек
			subTasks, err := ParseExpression(exprID, strings.Join(subExpr, ""))
			if err != nil {
				return nil, err
			}

			for _, task := range subTasks {
				*tasks = append(*tasks, task)
				result = append(result, task.ID)
			}
		} else {
			result = append(result, token)
		}
	}

	return result, nil
}

func (s *ExpressionService) GetAllExpressions() ([]*models.Expression, error) {
	return s.exprRepo.GetAll()
}

func (s *ExpressionService) GetExpressionByID(id string) (*models.Expression, error) {
	return s.exprRepo.GetByID(id)
}

func (s *ExpressionService) ProcessExpression(expression string) (string, error) {
	exprID := uuid.New().String()

	expr := &models.Expression{
		ID:     exprID,
		Status: "pending",
	}

	// Сохраняем выражение в БД
	if err := s.exprRepo.Create(expr); err != nil {
		return "", err
	}

	// Парсим выражение с учётом приоритетов операций и скобок
	tasks, err := ParseExpression(exprID, expression)
	if err != nil {
		return "", err
	}

	// Сохраняем задачи в БД
	for _, task := range tasks {
		if err := s.taskRepo.Create(task); err != nil {
			return "", err
		}
	}

	return exprID, nil
}

func (s *TaskService) GetDependenciesResults(task *models.Task) (float64, float64, error) {
	arg1, err := s.getOperandValue(task.Arg1)
	if err != nil {
		return 0, 0, err
	}

	arg2, err := s.getOperandValue(task.Arg2)
	if err != nil {
		return 0, 0, err
	}

	return arg1, arg2, nil
}

func (s *TaskService) getOperandValue(operand interface{}) (float64, error) {
	switch v := operand.(type) {
	case float64:
		return v, nil
	case string:
		task, err := s.repo.GetByID(v)
		if err != nil {
			return 0, err
		}
		return task.Result, nil
	default:
		return 0, fmt.Errorf("invalid operand type: %T", operand)
	}
}
