package calculator

import (
	"calculator_app/internal/orchestrator/models"
	"errors"
	"fmt"
	"github.com/google/uuid"
	//"strconv"
	"strings"
)

// CalculationStep представляет шаг вычисления
type CalculationStep struct {
	Arg1      string
	Arg2      string
	Operation string
	Result    string
}

// TaskBuilder — строитель задач
type TaskBuilder struct {
	expressionID string
	tasks        []*models.Task
	operandStack []string // Хранит ID предыдущих задач или числа
}

// Создаёт новый уникальный ID задачи
func (tb *TaskBuilder) newTaskID() string {
	return uuid.New().String()
}

// Парсит выражение в последовательность шагов вычислений
func ParseToSteps(expression string) ([]CalculationStep, error) {
	tokens, err := tokenize(expression)
	if err != nil {
		return nil, err
	}

	// Преобразуем выражение в обратную польскую нотацию (RPN)
	rpn, err := shuntingYard(tokens)
	if err != nil {
		return nil, err
	}

	var steps []CalculationStep
	var stack []string

	for _, token := range rpn {
		if isOperator(token) {
			if len(stack) < 2 {
				return nil, errors.New("неверный формат выражения")
			}

			// Берём два последних аргумента из стека
			arg2 := stack[len(stack)-1]
			arg1 := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			// Создаём ID для этой операции
			resultID := uuid.New().String()

			// Добавляем шаг вычисления
			steps = append(steps, CalculationStep{
				Arg1:      arg1,
				Arg2:      arg2,
				Operation: token,
				Result:    resultID,
			})

			// Кладём результат (taskID) обратно в стек
			stack = append(stack, resultID)
		} else {
			// Операнд просто кладём в стек
			stack = append(stack, token)
		}
	}

	if len(stack) != 1 {
		return nil, errors.New("ошибка обработки выражения")
	}

	return steps, nil
}

// Создаёт задачи на вычисления из выражения
func ParseToTasks(expression, expressionID string) ([]*models.Task, error) {
	tokens, err := tokenize(expression)
	if err != nil {
		return nil, err
	}

	postfix, err := shuntingYard(tokens)
	if err != nil {
		return nil, err
	}

	tb := &TaskBuilder{
		expressionID: expressionID,
		operandStack: make([]string, 0),
	}

	for _, token := range postfix {
		if isOperator(token) {
			if len(tb.operandStack) < 2 {
				return nil, errors.New("недостаточно операндов для оператора " + token)
			}

			// Берём два последних операнда
			arg2 := tb.operandStack[len(tb.operandStack)-1]
			arg1 := tb.operandStack[len(tb.operandStack)-2]
			tb.operandStack = tb.operandStack[:len(tb.operandStack)-2]

			// Создаём новую задачу
			task := &models.Task{
				ID:           tb.newTaskID(),
				ExpressionID: expressionID,
				Operation:    token,
				Arg1:         arg1,
				Arg2:         arg2,
				Status:       "pending",
			}

			// Результат этой задачи становится новым операндом
			tb.operandStack = append(tb.operandStack, task.ID)
			tb.tasks = append(tb.tasks, task)
		} else {
			// Числа просто добавляем в стек операндов
			tb.operandStack = append(tb.operandStack, token)
		}
	}

	if len(tb.operandStack) != 1 {
		return nil, errors.New("некорректное выражение")
	}

	return tb.tasks, nil
}

// Преобразует инфиксное выражение в RPN (обратную польскую нотацию)
func shuntingYard(tokens []string) ([]string, error) {
	var output []string
	var operators []string

	precedence := map[string]int{
		"+": 1, "-": 1,
		"*": 2, "/": 2,
	}

	for _, token := range tokens {
		if isOperator(token) {
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
			// Число добавляем в выходной список
			output = append(output, token)
		}
	}

	// Добавляем оставшиеся операторы
	for len(operators) > 0 {
		output = append(output, operators[len(operators)-1])
		operators = operators[:len(operators)-1]
	}

	return output, nil
}

// Разбивает строку-выражение на токены
func tokenize(expression string) ([]string, error) {
	var tokens []string
	var numberBuffer strings.Builder

	for _, ch := range expression {
		switch {
		case ch >= '0' && ch <= '9' || ch == '.': // Число
			numberBuffer.WriteRune(ch)
		case ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '(' || ch == ')':
			// Если перед оператором было число — добавляем его в токены
			if numberBuffer.Len() > 0 {
				tokens = append(tokens, numberBuffer.String())
				numberBuffer.Reset()
			}
			// Добавляем сам оператор
			tokens = append(tokens, string(ch))
		case ch == ' ': // Пропускаем пробелы
			if numberBuffer.Len() > 0 {
				tokens = append(tokens, numberBuffer.String())
				numberBuffer.Reset()
			}
		default:
			return nil, fmt.Errorf("некорректный символ: %c", ch)
		}
	}

	// Добавляем последнее число
	if numberBuffer.Len() > 0 {
		tokens = append(tokens, numberBuffer.String())
	}

	return tokens, nil
}

// Проверяет, является ли строка оператором
func isOperator(s string) bool {
	return s == "+" || s == "-" || s == "*" || s == "/"
}
