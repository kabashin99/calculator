package repository

import (
	"calculator_app/internal/pkg/models"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
)

const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	ExprStatusDone       = "done"
)

type Repository struct {
	db *sql.DB
}

type RepositoryInterface interface {
	AddExpression(expr *models.Expression) error
	AddTask(task *models.Task) error
	GetAndLockTask() (*models.Task, bool, error)
	UpdateTaskResult(taskID string, result *float64, taskErr *models.TaskError) (bool, string, error)
	UpdateExpression(id string, status string, result float64) (bool, error)
	CalculateFinalResult(expressionID string) (float64, error)
	AreAllTasksCompleted(expressionID string) (bool, error)
	GetExpressionsByOwner(owner string) (map[string]*models.Expression, error)
	GetExpressionByIDAndOwner(id, owner string) (*models.Expression, bool, error)
	RegisterUser(user models.User) error
	FindUser(login string) (*models.User, error)
	GetTaskResult(taskID string) (float64, bool, error)
}

var ErrUserExists = errors.New("user already exists")

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AddExpression(expr *models.Expression) error {
	var result interface{} = nil
	if expr.Result != nil {
		result = *expr.Result
	}

	_, err := r.db.Exec(
		`INSERT INTO expressions (id, status, result, owner) VALUES (?, ?, ?, ?)`,
		expr.ID, expr.Status, result, expr.Owner,
	)
	return err
}

func (r *Repository) AddTask(task *models.Task) error {
	var result interface{} = nil
	if task.Result != nil {
		result = *task.Result
	}

	dependsOn := strings.Join(task.DependsOn, ",")

	_, err := r.db.Exec(
		`INSERT INTO tasks 
			(id, arg1, arg2, operation, operation_time, result, depends_on, user_login) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.Arg1, task.Arg2, task.Operation, task.OperationTime,
		result, dependsOn, task.UserLogin,
	)
	if task.Status == "" {
		task.Status = TaskStatusPending
	}

	return err
}

func (r *Repository) RegisterUser(user models.User) error {
	_, err := r.db.Exec(
		`INSERT INTO users (login, password) VALUES (?, ?)`,
		user.Login, user.Password,
	)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: users.login") {
		return ErrUserExists
	}
	return err
}

func (r *Repository) FindUser(login string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(
		`SELECT login, password FROM users WHERE login = ?`,
		login,
	).Scan(&user.Login, &user.Password)

	if err != nil {
		return &models.User{}, err
	}

	return &user, err
}

func (r *Repository) GetTaskResult(taskID string) (float64, bool, error) {
	var result float64
	err := r.db.QueryRow(
		`SELECT result FROM tasks WHERE id = ? AND result IS NOT NULL AND result != 0`,
		taskID,
	).Scan(&result)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return result, true, nil
}

func (r *Repository) GetExpressionsByOwner(owner string) (map[string]*models.Expression, error) {
	rows, err := r.db.Query(
		`SELECT id, status, result, owner FROM expressions WHERE owner = ?`,
		owner,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("Warning: failed to close rows: %v", cerr)
		}
	}()

	expressions := make(map[string]*models.Expression)
	for rows.Next() {
		var expr models.Expression
		if err := rows.Scan(&expr.ID, &expr.Status, &expr.Result, &expr.Owner); err != nil {
			return nil, err
		}
		expressions[expr.ID] = &expr
	}
	return expressions, nil
}

func (r *Repository) GetExpressionByIDAndOwner(id string, owner string) (*models.Expression, bool, error) {
	var expr models.Expression
	err := r.db.QueryRow(
		`SELECT id, status, result, owner FROM expressions WHERE id = ? AND owner = ?`,
		id, owner,
	).Scan(&expr.ID, &expr.Status, &expr.Result, &expr.Owner)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &expr, true, nil
}

func (r *Repository) GetAndLockTask() (*models.Task, bool, error) {

	tx, err := r.db.Begin()
	if err != nil {
		return nil, false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rErr := tx.Rollback(); rErr != nil && !errors.Is(rErr, sql.ErrTxDone) {
			log.Printf("Warning: transaction rollback failed: %v", rErr)
		}
	}()

	var task models.Task
	var dependsOnStr string
	var result sql.NullFloat64

	err = tx.QueryRow(`
		SELECT id, arg1, arg2, operation, operation_time, depends_on, user_login, result
		FROM tasks 
		WHERE status = ? AND result IS NULL
		ORDER BY created_at ASC
		LIMIT 1`,
		TaskStatusPending,
	).Scan(
		&task.ID, &task.Arg1, &task.Arg2, &task.Operation,
		&task.OperationTime, &dependsOnStr, &task.UserLogin, &result,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("query error: %w", err)
	}

	if result.Valid {
		task.Result = &result.Float64
	} else {
		task.Result = nil
	}

	res, err := tx.Exec(`
        UPDATE tasks 
        SET status = ?, 
            updated_at = CURRENT_TIMESTAMP
        WHERE id = ? 
          AND status = ?`,
		TaskStatusProcessing, task.ID, TaskStatusPending)
	if err != nil {
		return nil, false, fmt.Errorf("update error: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, false, fmt.Errorf("rows affected error: %w", err)
	}

	if rowsAffected == 0 {
		return nil, false, nil
	}

	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit error: %w", err)
	}

	if dependsOnStr != "" {
		task.DependsOn = strings.Split(dependsOnStr, ",")
	}

	task.Status = TaskStatusProcessing

	return &task, true, nil
}

func (r *Repository) UpdateTaskResult(taskID string, result *float64, taskErr *models.TaskError) (bool, string, error) {
	var (
		status      string
		resultValue sql.NullFloat64
		//errorMessage sql.NullString
	)
	if taskErr != nil {
		status = string(taskErr.Code)
		resultValue = sql.NullFloat64{}
		//errorMessage = sql.NullString{
		//	String: taskErr.Message,
		//	Valid:  true,
		//}
	} else {
		status = TaskStatusCompleted
		resultValue = sql.NullFloat64{
			Float64: *result,
			Valid:   true,
		}
		//errorMessage = sql.NullString{} // нет ошибки — поле пустое
	}

	res, err := r.db.Exec(
		`UPDATE tasks SET 
            result = ?, 
            status = ?,
            updated_at = CURRENT_TIMESTAMP
         WHERE id = ? AND status = ?`,
		resultValue,
		status,
		taskID,
		TaskStatusProcessing,
	)
	if err != nil {
		return false, "", fmt.Errorf("failed to update task result: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, "", fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected > 0, status, nil
}

func (r *Repository) AreAllTasksCompleted(exprID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM tasks 
         WHERE id LIKE ? || '-%' AND status != ? `,
		exprID, TaskStatusCompleted,
	).Scan(&count)

	return count == 0, err
}

func (r *Repository) CalculateFinalResult(exprID string) (float64, error) {
	var result float64
	err := r.db.QueryRow(
		`
        SELECT t.result
        FROM tasks AS t
        WHERE t.id LIKE ? || '-%'
          AND t.status = ?
          AND NOT EXISTS (
              SELECT 1
              FROM tasks AS t2
              WHERE t2.id LIKE ? || '-%'
                AND t2.depends_on LIKE '%' || t.id || '%'
          )
        ORDER BY LENGTH(t.depends_on) DESC
        LIMIT 1;`,
		exprID, TaskStatusCompleted, exprID,
	).Scan(&result)

	return result, err
}

func (r *Repository) UpdateExpression(exprID string, status string, result float64) (bool, error) {

	if status == TaskStatusCompleted {
		status = ExprStatusDone
	}

	res, err := r.db.Exec(
		`UPDATE expressions 
			   SET status = ?, result = ? 
			   WHERE id = ?`,
		status, result, exprID,
	)

	if err != nil {
		return false, fmt.Errorf("failed to execute update: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return rowsAffected > 0, nil
}
