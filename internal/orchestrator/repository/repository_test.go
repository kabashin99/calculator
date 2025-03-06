package repository

import (
	"testing"

	"calculator_app/internal/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAddExpression(t *testing.T) {
	repo := NewRepository()
	expr := &models.Expression{ID: "expr-1", Status: "pending", Result: 0}

	repo.AddExpression(expr)

	exprFromRepo, exists := repo.GetExpressionByID("expr-1")
	assert.True(t, exists)
	assert.Equal(t, expr, exprFromRepo)
}

func TestGetExpressions(t *testing.T) {
	repo := NewRepository()
	expr1 := &models.Expression{ID: "expr-1", Status: "done", Result: 42}
	expr2 := &models.Expression{ID: "expr-2", Status: "pending", Result: 0}

	repo.AddExpression(expr1)
	repo.AddExpression(expr2)

	expressions := repo.GetExpressions()
	assert.Len(t, expressions, 2)
	assert.Equal(t, expr1, expressions["expr-1"])
	assert.Equal(t, expr2, expressions["expr-2"])
}

func TestGetExpressionByID(t *testing.T) {
	repo := NewRepository()
	expr := &models.Expression{ID: "expr-1", Status: "pending", Result: 0}

	repo.AddExpression(expr)

	exprFromRepo, exists := repo.GetExpressionByID("expr-1")
	assert.True(t, exists)
	assert.Equal(t, expr, exprFromRepo)

	_, exists = repo.GetExpressionByID("unknown-id")
	assert.False(t, exists)
}

func TestAddTask(t *testing.T) {
	repo := NewRepository()
	task := &models.Task{ID: "task-1", Arg1: 5, Arg2: 3, Operation: "+"}

	repo.AddTask(task)

	taskFromRepo, exists := repo.GetTask()
	assert.True(t, exists)
	assert.Equal(t, task, taskFromRepo)
}

func TestGetTask(t *testing.T) {
	repo := NewRepository()
	task1 := &models.Task{ID: "task-1", Arg1: 2, Arg2: 3, Operation: "*"}
	task2 := &models.Task{ID: "task-2", Arg1: 10, Arg2: 5, Operation: "-"}

	repo.AddTask(task1)
	repo.AddTask(task2)

	taskFromRepo, exists := repo.GetTask()
	assert.True(t, exists)
	assert.Equal(t, task1, taskFromRepo)

	taskFromRepo, exists = repo.GetTask()
	assert.True(t, exists)
	assert.Equal(t, task2, taskFromRepo)

	_, exists = repo.GetTask()
	assert.False(t, exists)
}

func TestUpdateExpressionResult(t *testing.T) {
	repo := NewRepository()
	expr := &models.Expression{ID: "expr-1", Status: "pending", Result: 0}
	repo.AddExpression(expr)

	success := repo.UpdateExpressionResult("expr-1", 99.9)
	assert.True(t, success)

	updatedExpr, exists := repo.GetExpressionByID("expr-1")
	assert.True(t, exists)
	assert.Equal(t, 99.9, updatedExpr.Result)
	assert.Equal(t, "done", updatedExpr.Status)

	success = repo.UpdateExpressionResult("unknown-id", 50)
	assert.False(t, success)
}
