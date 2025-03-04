package utils

import (
	"testing"
)

func TestIsValidExpression(t *testing.T) {
	testCases := []struct {
		expression string
		expected   bool
	}{
		{"1 + 2", true},
		{"10 * 5", true},
		{"(1 + 2) * 3", true},
		{"1 + 2 + 3", true},
		{"1 + a", false},
		{"1 + 2 + 3", true},
		{"1++2", true},
		{"1/0", true},
		{"", true},
		{"(", true},
		{"1 + 2.5", false},
		{"(1+2)", true},
		{"1 + (2 * 3)", true},
	}

	for _, tc := range testCases {
		actual := IsValidExpression(tc.expression)
		if actual != tc.expected {
			t.Errorf("IsValidExpression(%q) = %t; want %t", tc.expression, actual, tc.expected)
		}
	}
}

func TestIsOperator(t *testing.T) {
	testCases := []struct {
		c        string
		expected bool
	}{
		{"+", true},
		{"-", true},
		{"*", true},
		{"/", true},
		{"", false},
		{"a", false},
		{"5", false},
		{"(", false},
		{")", false},
	}

	for _, tc := range testCases {
		actual := IsOperator(tc.c)
		if actual != tc.expected {
			t.Errorf("IsOperator(%q) = %t; want %t", tc.c, actual, tc.expected)
		}
	}
}

func TestIsDigit(t *testing.T) {
	testCases := []struct {
		c        string
		expected bool
	}{
		{"123", true},
		{"3.14", true},
		{"-5", true},
		{"0", true},
		{"0.0", true},
		{" ", false},
		{"abc", false},
		{"1a", false},
		{"1.a", false},
		{".1", true},
		{"-1.23", true},
		{"+1.23", true},
	}

	for _, tc := range testCases {
		actual := IsDigit(tc.c)
		if actual != tc.expected {
			t.Errorf("IsDigit(%q) = %t; want %t", tc.c, actual, tc.expected)
		}
	}
}
