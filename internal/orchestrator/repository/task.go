package repository

import (
	"calculator_app/internal/orchestrator/models"
	"database/sql"
	"sync"
)

// Определяем интерфейс
type TaskRepository interface {
	Create(task *models.Task) error
	GetNextPending() (*models.Task, error)
	UpdateStatus(id string, status string, result float64) error
	GetAll() ([]*models.Task, error)
	GetByID(id string) (*models.Task, error)
	GetByExpressionID(exprID string) ([]*models.Task, error) // <-- Добавлен метод
}

// Структура с мьютексом
type taskRepoImpl struct {
	db    *sql.DB
	mu    sync.RWMutex
	tasks map[string]*models.Task
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	db.SetMaxOpenConns(1) // SQLite не поддерживает параллельную запись, ограничиваем потоки
	return &taskRepoImpl{
		db:    db,
		tasks: make(map[string]*models.Task),
	}
}

func (r *taskRepoImpl) Create(task *models.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, err := r.db.Exec("INSERT INTO tasks (id, expression_id, arg1, arg2, operation, status) VALUES (?, ?, ?, ?, ?, ?)",
		task.ID, task.ExpressionID, task.Arg1, task.Arg2, task.Operation, task.Status)
	if err == nil {
		r.tasks[task.ID] = task
	}
	return err
}

func (r *taskRepoImpl) GetNextPending() (*models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	row := r.db.QueryRow("SELECT id, expression_id, arg1, arg2, operation, status FROM tasks WHERE status = 'pending' LIMIT 1")

	task := &models.Task{}
	err := row.Scan(&task.ID, &task.ExpressionID, &task.Arg1, &task.Arg2, &task.Operation, &task.Status)
	return task, err
}

func (r *taskRepoImpl) UpdateStatus(id string, status string, result float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec("UPDATE tasks SET status = ?, result = ? WHERE id = ?", status, result, id)
	return err
}

func (r *taskRepoImpl) GetAll() ([]*models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rows, err := r.db.Query("SELECT id, expression_id, arg1, arg2, operation, status FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(&task.ID, &task.ExpressionID, &task.Arg1, &task.Arg2, &task.Operation, &task.Status)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *taskRepoImpl) GetByID(id string) (*models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task := &models.Task{}
	err := r.db.QueryRow("SELECT id, expression_id, arg1, arg2, operation, status FROM tasks WHERE id = ?", id).
		Scan(&task.ID, &task.ExpressionID, &task.Arg1, &task.Arg2, &task.Operation, &task.Status)
	return task, err
}

func (r *taskRepoImpl) GetByExpressionID(exprID string) ([]*models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rows, err := r.db.Query("SELECT id, expression_id, arg1, arg2, operation, status, result FROM tasks WHERE expression_id = ?", exprID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(&task.ID, &task.ExpressionID, &task.Arg1, &task.Arg2, &task.Operation, &task.Status, &task.Result)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
