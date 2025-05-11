package repository_test

import (
	"database/sql"
	"testing"

	"calculator_app/internal/orchestrator/repository"
	"calculator_app/internal/pkg/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// setupMock включает Regexp-матчер для всех ExpectExec/ExpectQuery
func setupMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	return db, mock
}

func TestAddExpression(t *testing.T) {
	db, mock := setupMock(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	expr := &models.Expression{
		ID:     "expr123",
		Status: "pending",
		Result: nil,
		Owner:  "user1",
	}

	// Регексп, матчущий начало INSERT
	mock.ExpectExec(`^INSERT INTO expressions`).
		WithArgs(expr.ID, expr.Status, nil, expr.Owner).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.AddExpression(expr)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddTask(t *testing.T) {
	db, mock := setupMock(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	task := &models.Task{
		ID:            "task1",
		Arg1:          1.1,
		Arg2:          2.2,
		Operation:     "+",
		OperationTime: 5,
		Result:        nil,
		DependsOn:     []string{"task0", "taskX"},
		UserLogin:     "user1",
		Status:        "",
	}

	// Простая регулярка по префиксу
	mock.ExpectExec(`^INSERT INTO tasks`).
		WithArgs(
			task.ID,
			task.Arg1,
			task.Arg2,
			task.Operation,
			task.OperationTime,
			nil,
			"task0,taskX",
			task.UserLogin,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.AddTask(task)
	assert.NoError(t, err)
	assert.Equal(t, repository.TaskStatusPending, task.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegisterUser(t *testing.T) {
	db, mock := setupMock(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	user := models.User{
		Login:    "testuser",
		Password: "securepassword",
	}

	mock.ExpectExec(`^INSERT INTO users`).
		WithArgs(user.Login, user.Password).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.RegisterUser(user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindUser(t *testing.T) {
	db, mock := setupMock(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	login := "user1"
	expectedUser := models.User{
		Login:    login,
		Password: "pass123",
	}
	rows := sqlmock.NewRows([]string{"login", "password"}).
		AddRow(expectedUser.Login, expectedUser.Password)

	mock.ExpectQuery(`^SELECT login, password FROM users`).
		WithArgs(login).
		WillReturnRows(rows)

	user, err := repo.FindUser(login)
	assert.NoError(t, err)
	assert.Equal(t, &expectedUser, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTaskResult_WithResult(t *testing.T) {
	db, mock := setupMock(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	taskID := "task1"
	expectedResult := 42.0

	rows := sqlmock.NewRows([]string{"result"}).
		AddRow(expectedResult)

	mock.ExpectQuery(`^SELECT result FROM tasks`).
		WithArgs(taskID).
		WillReturnRows(rows)

	result, ok, err := repo.GetTaskResult(taskID)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, expectedResult, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetExpressionByIDAndOwner_Found(t *testing.T) {
	db, mock := setupMock(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	id, owner := "expr123", "user1"
	expectedVal := 3.14

	rows := sqlmock.NewRows([]string{"id", "status", "result", "owner"}).
		AddRow(id, "done", expectedVal, owner)

	mock.ExpectQuery(`^SELECT id, status, result, owner FROM expressions`).
		WithArgs(id, owner).
		WillReturnRows(rows)

	expr, found, err := repo.GetExpressionByIDAndOwner(id, owner)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, id, expr.ID)
	assert.Equal(t, "done", expr.Status)
	assert.NotNil(t, expr.Result)
	assert.Equal(t, expectedVal, *expr.Result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
