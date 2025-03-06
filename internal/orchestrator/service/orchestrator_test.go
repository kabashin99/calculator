package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddExpression(t *testing.T) {
	orc := NewOrchestrator(100, 100, 200, 200)

	id, err := orc.AddExpression("3 + 5 * 2")
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	expr, exists := orc.GetExpressionByID(id)
	assert.True(t, exists)
	assert.Equal(t, "pending", expr.Status)
	assert.Equal(t, 0.0, expr.Result)

	assert.Greater(t, len(orc.tasks), 0)
}

func TestGetExpressions(t *testing.T) {
	orc := NewOrchestrator(100, 100, 200, 200)

	id1, _ := orc.AddExpression("10 + 2")
	id2, _ := orc.AddExpression("5 * 3")

	expressions := orc.GetExpressions()
	assert.Len(t, expressions, 2)
	assert.Contains(t, expressions, id1)
	assert.Contains(t, expressions, id2)
}

func TestGetExpressionByID(t *testing.T) {
	orc := NewOrchestrator(100, 100, 200, 200)

	id, _ := orc.AddExpression("8 / 2")
	expr, exists := orc.GetExpressionByID(id)

	assert.True(t, exists)
	assert.Equal(t, "pending", expr.Status)

	_, exists = orc.GetExpressionByID("unknown-id")
	assert.False(t, exists)
}

func TestGetTask(t *testing.T) {
	orc := NewOrchestrator(100, 100, 200, 200)

	orc.AddExpression("4 + 6")

	task, exists := orc.GetTask()
	assert.True(t, exists)
	assert.NotNil(t, task)
	assert.NotEmpty(t, task.ID)

	_, exists = orc.GetTask()
	assert.False(t, exists)
}

func TestSubmitResult(t *testing.T) {
	orc := NewOrchestrator(100, 100, 200, 200)

	id, _ := orc.AddExpression("5 + 3")
	task, _ := orc.GetTask()

	success := orc.SubmitResult(task.ID, 8.0)
	assert.True(t, success)

	expr, exists := orc.GetExpressionByID(id)
	assert.True(t, exists)
	assert.Equal(t, 8.0, expr.Result)
	assert.Equal(t, "done", expr.Status)
}

func TestSubmitResult_InvalidTaskID(t *testing.T) {
	orc := NewOrchestrator(100, 100, 200, 200)

	success := orc.SubmitResult("invalid-task", 42.0)
	assert.False(t, success)
}

func TestParseExpressionToTasks(t *testing.T) {
	orc := NewOrchestrator(100, 100, 200, 200)

	tasks, err := orc.parseExpressionToTasks("3 + 5 * 2", "expr-1")
	assert.NoError(t, err)
	assert.Len(t, tasks, 2) // Должно быть 2 задачи (умножение имеет приоритет)

	assert.Equal(t, "+", tasks[0].Operation)
	assert.Equal(t, "*", tasks[1].Operation)
}

func TestTokenize(t *testing.T) {
	tokens := tokenize("3 + 5 * 2")
	expected := []string{"3", "+", "5", "*", "2"}
	assert.Equal(t, expected, tokens)
}

func TestShuntingYard(t *testing.T) {
	tokens := []string{"3", "+", "5", "*", "2"}
	postfix, err := shuntingYard(tokens)
	assert.NoError(t, err)

	expected := []string{"3", "5", "2", "*", "+"}
	assert.Equal(t, expected, postfix)
}
