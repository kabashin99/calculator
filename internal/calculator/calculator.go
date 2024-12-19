// Надо переделать калькулятор в соответствии с изменениями в других файлах.

package calculator

import (
	"calculator_app/internal/utils"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Stack []string

// Добавление элемента в стек
func (s *Stack) Push(value string) {
	*s = append(*s, value)
}

// извлекает и возвращает верхний элемент стека. Если стек пуст, он возвращает пустую строку.
func (s *Stack) Pop() string {
	if len(*s) == 0 {
		return ""
	}
	value := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return value
}

// получение верхнего элемента стека без его удаления
func (s *Stack) Top() string {
	if len(*s) == 0 {
		return ""
	}
	return (*s)[len(*s)-1]
}

// позволяет определить приоритет математических операторов
func precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	}
	return 0
}

// преобразование математических выражений
func shuntingYard(expression string) ([]string, error) {
	var output []string
	var operators Stack

	// tokens := strings.Fields(expression)
	// Разбиваем выражение на токены, учитывая пробелы и скобки
	tokens := splitExpression(expression)

	for _, token := range tokens {
		if utils.IsDigit(token) {
			output = append(output, token)
		} else if token == "(" {
			operators.Push(token)
		} else if token == ")" {
			if len(operators) == 0 {
				return nil, errors.New("mismatched parentheses")
			}
			for operators.Top() != "(" {
				output = append(output, operators.Pop())
				if len(operators) == 0 {
					return nil, errors.New("mismatched parentheses")
				}
			}
			operators.Pop() // Удаляем "("
		} else if utils.IsOperator(token) {
			for len(operators) > 0 && precedence(operators.Top()) >= precedence(token) {
				output = append(output, operators.Pop())
			}
			operators.Push(token)
		} else {
			return nil, errors.New("invalid token: " + token)
		}
	}

	for len(operators) > 0 {
		if operators.Top() == "(" {
			return nil, errors.New("mismatched parentheses")
		}
		output = append(output, operators.Pop())
	}

	return output, nil
}

// Функция для разделения выражения на токены, учитывая пробелы и скобки
func splitExpression(expression string) []string {
	var tokens []string
	currentToken := ""
	for _, char := range expression {
		if char == ' ' || char == '(' || char == ')' {
			if currentToken != "" {
				tokens = append(tokens, currentToken)
				currentToken = ""
			}
			if char != ' ' {
				tokens = append(tokens, string(char))
			}
		} else {
			currentToken += string(char)
		}
	}
	if currentToken != "" {
		tokens = append(tokens, currentToken)
	}
	return tokens
}

// функция вычисляет результаты математических выражений
func evaluatePostfix(postfix []string) (float64, error) {
	var stack Stack

	for _, token := range postfix {
		if utils.IsDigit(token) {
			stack.Push(token)
		} else if utils.IsOperator(token) {
			if len(stack) < 2 {
				return 0, errors.New("invalid expression: not enough operands")
			}
			b := stack.Pop()
			a := stack.Pop()

			numA, errA := strconv.ParseFloat(a, 64)
			numB, errB := strconv.ParseFloat(b, 64)

			if errA != nil || errB != nil {
				return 0, errors.New("invalid number format")
			}

			var result float64
			switch token {
			case "+":
				result = numA + numB
			case "-":
				result = numA - numB
			case "*":
				result = numA * numB
			case "/":
				if numB == 0 {
					return 0, errors.New("division by zero")
				}
				result = numA / numB
			default:
				return 0, errors.New("unknown operator: " + token)
			}
			stack.Push(fmt.Sprintf("%f", result))
		} else {
			return 0, errors.New("invalid token: " + token)
		}
	}

	if len(stack) != 1 {
		return 0, errors.New("invalid expression: too many operands remaining")
	}

	resultStr := stack.Pop()
	result, err := strconv.ParseFloat(resultStr, 64)
	if err != nil {
		return 0, errors.New("invalid number format")
	}
	return result, nil
}

func addSpaceAfterChars(str string) string {
	result := ""
	for i, char := range str {
		result += string(char)
		if i < len(str)-1 { // Не добавляем пробел после последнего символа
			result += " "
		}
	}
	return result
}

func Calc(expression string) (float64, error) {

	// Удаляем пробелы из выражения
	expression = addSpaceAfterChars(strings.ReplaceAll(expression, " ", ""))

	// Проверяем, является ли выражение пустым
	if expression == "" {
		return 0, errors.New("Expression is empty")
	}

	// Проверяем наличие скобок
	openBracketsCount := strings.Count(expression, "(")
	closeBracketsCount := strings.Count(expression, ")")
	if openBracketsCount != closeBracketsCount {
		return 0, errors.New("Brackets are not balanced")
	}

	postfix, err := shuntingYard(expression)
	if err != nil {
		return 0, err
	}
	return evaluatePostfix(postfix)
}
