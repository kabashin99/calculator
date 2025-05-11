package models_test

import (
	"calculator_app/internal/pkg/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTaskErrorAndErrorString(t *testing.T) {
	te := models.NewTaskError(models.ErrDivisionByZero, "cannot divide by zero")

	assert.Equal(t, models.ErrDivisionByZero, te.Code)
	assert.Equal(t, "cannot divide by zero", te.Message)

	expected := "division_by_zero: cannot divide by zero"
	assert.Equal(t, expected, te.Error())
}
