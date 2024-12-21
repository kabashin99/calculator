package utils

import (
	"strconv"
	"strings"
)

// Проверка валидности выражения
func IsValidExpression(expression string) bool {
	for _, char := range expression {
		if !((char >= '0' && char <= '9') ||
			char == '+' ||
			char == '-' ||
			char == '*' ||
			char == '/' ||
			char == '(' ||
			char == ')' ||
			char == ' ') {
			return false
		}
	}
	return true
}

func IsOperator(c string) bool {
	if c == "" {
		return false
	}

	operators := "+-*/"
	return strings.Contains(operators, c)
}

func IsDigit(c string) bool {
	if strings.TrimSpace(c) == "" {
		return false
	}

	_, err := strconv.ParseFloat(c, 64)
	return err == nil
}
