package repository

import (
	"calculator_app/internal/pkg/models"
	"database/sql"
	"github.com/lib/pq"
	"log"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Создание выражения
func (r *Repository) AddExpression(expression *models.Expression) error {
	query := `
		INSERT INTO expressions (id, user_id, status, result) 
		VALUES ($1, $2, $3, $4)
		`
	_, err := r.db.Exec(query, expression.ID, expression.UserID, expression.Status, expression.Result)
	if err != nil {
		log.Printf("Error inserting expression: %v", err)
	}
	return err
}

// Получение выражения по ID
func (r *Repository) GetExpressionByID(id string) (*models.Expression, error) {
	query := `SELECT id, user_id, status, result FROM expressions WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var expr models.Expression
	err := row.Scan(&expr.ID, &expr.UserID, &expr.Status, &expr.Result)
	if err != nil {
		return nil, err
	}
	return &expr, nil
}

// Получение всех выражений пользователя
func (r *Repository) GetExpressionsByUserID(userID string) ([]*models.Expression, error) {
	query := `SELECT id, user_id, status, result FROM expressions WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expressions []*models.Expression
	for rows.Next() {
		var expr models.Expression
		err := rows.Scan(&expr.ID, &expr.UserID, &expr.Status, &expr.Result)
		if err != nil {
			log.Printf("Error scanning expression: %v", err)
			continue
		}
		expressions = append(expressions, &expr)
	}

	return expressions, nil
}

// Обновление результата выражения
func (r *Repository) UpdateExpressionResult(id string, result float64) error {
	query := `
		UPDATE tasks
		SET result = $1, status = 'done'
		WHERE id = $2
	`

	_, err := r.db.Exec(query, result, id)
	return err
}

// Добавление задачи
func (r *Repository) AddTask(task *models.Task) error {
	query := `
		INSERT INTO tasks (id, expression_id, operation, operand1, operand2, operation_time, result, depends_on, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(query,
		task.ID,
		task.ExpressionID,
		task.Operation,
		task.Arg1,
		task.Arg2,
		task.OperationTime,
		task.Result,
		pq.Array(task.DependsOn),
		task.Status,
	)

	return err
}

// Получение первой "свободной" задачи
func (r *Repository) GetTask() (*models.Task, error) {
	query := `
		SELECT id, expression_id, operation, operand1, operand2, operation_time, result, depends_on, status
		FROM tasks
		WHERE status = 'pending'
		ORDER BY operation_time ASC
		LIMIT 1
	`

	row := r.db.QueryRow(query)

	var task models.Task
	var dependsOn []string

	err := row.Scan(
		&task.ID,
		&task.ExpressionID,
		&task.Operation,
		&task.Arg1,
		&task.Arg2,
		&task.OperationTime,
		&task.Result,
		pq.Array(&dependsOn),
		&task.Status,
	)

	if err != nil {
		return nil, err
	}

	task.DependsOn = dependsOn
	return &task, nil
}

func (r *Repository) UpdateTaskResult(taskID string, result float64) error {
	// Сначала получаем связанное выражение
	var expressionID string
	err := r.db.QueryRow(`SELECT expression_id FROM tasks WHERE id = $1`, taskID).Scan(&expressionID)
	if err != nil {
		return err
	}

	// Обновляем задачу
	_, err = r.db.Exec(`UPDATE tasks SET status = 'done' WHERE id = $1`, taskID)
	if err != nil {
		return err
	}

	// Обновляем результат выражения
	_, err = r.db.Exec(`UPDATE expressions SET status = 'done', result = $1 WHERE id = $2`, result, expressionID)
	return err
}

/*
package repository

import (
	"calculator_app/internal/pkg/models"
	"sync"
)

type Repository struct {
	expressions map[string]*models.Expression
	tasks       []*models.Task
	users       map[string]*models.User
	mu          sync.Mutex
}

func NewRepository() *Repository {
	return &Repository{
		expressions: make(map[string]*models.Expression),
		tasks:       make([]*models.Task, 0),
		users:       make(map[string]*models.User),
	}
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

func (r *Repository) AddUser(user *models.User) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
}

func (r *Repository) GetUserByID(id string) (*models.User, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, exists := r.users[id]
	return user, exists
}
*/
