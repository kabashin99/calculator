package repository

import (
	"calculator_app/internal/orchestrator/models"
	"database/sql"
	"sync"
)

type ExpressionRepository struct {
	db    *sql.DB
	mu    sync.RWMutex
	exprs map[string]*models.Expression
}

func NewExpressionRepository(db *sql.DB) *ExpressionRepository {
	return &ExpressionRepository{
		db:    db,
		exprs: make(map[string]*models.Expression),
	}
}

func (r *ExpressionRepository) Create(expr *models.Expression) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec(
		"INSERT INTO expressions (id, status, result) VALUES (?, ?, ?)",
		expr.ID,
		expr.Status,
		expr.Result,
	)
	if err != nil {
		return err
	}

	r.exprs[expr.ID] = expr
	return nil
}

func (r *ExpressionRepository) GetByID(id string) (*models.Expression, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	expr := &models.Expression{}
	err := r.db.QueryRow(
		"SELECT id, status, result FROM expressions WHERE id = ?",
		id,
	).Scan(&expr.ID, &expr.Status, &expr.Result)

	return expr, err
}

func (r *ExpressionRepository) GetAll() ([]*models.Expression, error) {
	rows, err := r.db.Query("SELECT id, status, result FROM expressions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exprs []*models.Expression
	for rows.Next() {
		expr := &models.Expression{}
		if err := rows.Scan(&expr.ID, &expr.Status, &expr.Result); err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}
	return exprs, nil
}
