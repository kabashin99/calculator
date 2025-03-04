package repository

import (
	"calculator_app/internal/orchestrator/models"
	"database/sql"
	"sync"
)

type TaskRepository struct {
	db    *sql.DB
	mu    sync.RWMutex
	tasks map[string]*models.Task
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{
		db:    db,
		tasks: make(map[string]*models.Task),
	}
}

func (r *TaskRepository) Create(task *models.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec(
		"INSERT INTO tasks (id, expression_id, arg1, arg2, operation, status) VALUES (?, ?, ?, ?, ?, ?)",
		task.ID,
		task.ExpressionID,
		task.Arg1,
		task.Arg2,
		task.Operation,
		task.Status,
	)

	if err == nil {
		r.tasks[task.ID] = task
	}
	return err
}

func (r *TaskRepository) GetNextPending() (*models.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	row := r.db.QueryRow(
		"SELECT id, expression_id, arg1, arg2, operation FROM tasks WHERE status = 'pending' LIMIT 1",
	)

	task := &models.Task{}
	err := row.Scan(
		&task.ID,
		&task.ExpressionID,
		&task.Arg1,
		&task.Arg2,
		&task.Operation,
	)
	return task, err
}

func (r *TaskRepository) UpdateStatus(id string, status string) error {
	_, err := r.db.Exec(
		"UPDATE tasks SET status = ? WHERE id = ?",
		status,
		id,
	)
	return err
}
