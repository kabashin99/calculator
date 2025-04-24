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
func (r *Repository) AddExpression(expr *models.Expression) error {
	query := `INSERT INTO expressions (id, user_id, status, result)
		VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, expr.ID, expr.UserID, expr.Status, expr.Result)
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
		UPDATE expressions
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
		task.Operand1,
		task.Operand2,
		task.OperationTime,
		task.Result,
		pq.Array(task.DependsOn),
		task.Status,
	)

	return err
}

// Получение первой "свободной" задачи
func (r *Repository) GetTask() (*models.Task, error) {
	log.Println("[repository.go] GetTask called")

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
		&task.Operand1,
		&task.Operand2,
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
	// _, err = r.db.Exec(`UPDATE tasks SET status = 'done' WHERE id = $1`, taskID)
	_, err = r.db.Exec(`UPDATE tasks SET result = $1, status = 'done' WHERE id = $2`, result, taskID)
	if err != nil {
		return err
	}

	// Обновляем результат выражения
	//_, err = r.db.Exec(`UPDATE expressions SET status = 'done', result = $1 WHERE id = $2`, result, expressionID)
	return err
}

func (r *Repository) GetDB() *sql.DB {
	return r.db
}
