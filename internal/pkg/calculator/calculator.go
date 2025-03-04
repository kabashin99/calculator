package calculator

import (
	"calculator_app/internal/pkg/utils"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Stack []string

func (s *Stack) Push(value string) {
	*s = append(*s, value)
}

func (s *Stack) Pop() string {
	if len(*s) == 0 {
		return ""
	}
	value := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return value
}

func (s *Stack) Top() string {
	if len(*s) == 0 {
		return ""
	}
	return (*s)[len(*s)-1]
}

func precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	}
	return 0
}

func shuntingYard(expression string) ([]string, error) {
	var output []string
	var operators Stack

	tokens := splitExpression(expression)

	for _, token := range tokens {
		if utils.IsDigit(token) {
			output = append(output, token)
		} else if token == "(" {
			operators.Push(token)
		} else if token == ")" {
			if len(operators) == 0 {
				return nil, errors.New("несоответствующие скобки")
			}
			for operators.Top() != "(" {
				output = append(output, operators.Pop())
				if len(operators) == 0 {
					return nil, errors.New("несоответствующие скобки")
				}
			}
			operators.Pop()
		} else if utils.IsOperator(token) {
			for len(operators) > 0 && precedence(operators.Top()) >= precedence(token) {
				output = append(output, operators.Pop())
			}
			operators.Push(token)
		} else {
			return nil, errors.New("неверный токен: " + token)
		}
	}

	for len(operators) > 0 {
		if operators.Top() == "(" {
			return nil, errors.New("несоответствующие скобки")
		}
		output = append(output, operators.Pop())
	}
	return output, nil
}

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

func evaluatePostfix(postfix []string) (float64, error) {
	var stack Stack

	for _, token := range postfix {
		if utils.IsDigit(token) {
			stack.Push(token)
		} else if utils.IsOperator(token) {
			if len(stack) < 2 {
				return 0, errors.New("недопустимое выражение: недостаточно операндов")
			}
			b := stack.Pop()
			a := stack.Pop()

			numA, errA := strconv.ParseFloat(a, 64)
			numB, errB := strconv.ParseFloat(b, 64)

			if errA != nil || errB != nil {
				return 0, errors.New("неверный формат числа")
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
					return 0, errors.New("деление на ноль")
				}
				result = numA / numB
			default:
				return 0, errors.New("неизвестный оператор: " + token)
			}
			stack.Push(fmt.Sprintf("%f", result))
		} else {
			return 0, errors.New("неверный токен: " + token)
		}
	}

	if len(stack) != 1 {
		return 0, errors.New("недопустимое выражение: осталось слишком много операндов")
	}

	resultStr := stack.Pop()
	result, err := strconv.ParseFloat(resultStr, 64)
	if err != nil {
		return 0, errors.New("неверный формат числа")
	}
	return result, nil
}

func addSpaceAfterChars(str string) string {
	result := ""
	for i, char := range str {
		result += string(char)
		if i < len(str)-1 {
			result += " "
		}
	}
	return result
}

func Calc(expression string) (float64, error) {

	expression = addSpaceAfterChars(strings.ReplaceAll(expression, " ", ""))

	if expression == "" {
		return 0, errors.New("выражение пустое")
	}

	openBracketsCount := strings.Count(expression, "(")
	closeBracketsCount := strings.Count(expression, ")")
	if openBracketsCount != closeBracketsCount {
		return 0, errors.New("скобки не сбалансированы")
	}

	postfix, err := shuntingYard(expression)
	if err != nil {
		return 0, err
	}
	return evaluatePostfix(postfix)
}
