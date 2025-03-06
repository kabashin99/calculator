package repository

import (
	"calculator_app/internal/pkg/models"
	"sync"
)

type Repository struct {
	expressions map[string]*models.Expression
	tasks       []*models.Task
	mu          sync.Mutex
}

func (r *Repository) AddExpression(expression *models.Expression) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.expressions[expression.ID] = expression
}

func (r *Repository) GetExpressions() map[string]*models.Expression {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.expressions
}

func (r *Repository) GetExpressionByID(id string) (*models.Expression, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	expr, exists := r.expressions[id]
	return expr, exists
}

func (r *Repository) AddTask(task *models.Task) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks = append(r.tasks, task)
}

func (r *Repository) GetTask() (*models.Task, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.tasks) == 0 {
		return nil, false
	}

	task := r.tasks[0]
	r.tasks = r.tasks[1:]
	return task, true
}

func (r *Repository) UpdateExpressionResult(id string, result float64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	expr, exists := r.expressions[id]
	if !exists {
		return false
	}

	expr.Result = result
	expr.Status = "done"
	return true
}
