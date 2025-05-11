package service_test

import (
	"calculator_app/internal/orchestrator/service"
	"calculator_app/internal/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// Мокаем методы репозитория
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) AddExpression(expr *models.Expression) error {
	args := m.Called(expr)
	return args.Error(0)
}

func (m *MockRepository) AddTask(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockRepository) GetAndLockTask() (*models.Task, bool, error) {
	args := m.Called()
	return args.Get(0).(*models.Task), args.Bool(1), args.Error(2)
}

func (m *MockRepository) UpdateTaskResult(taskID string, result *float64, taskErr *models.TaskError) (bool, string, error) {
	args := m.Called(taskID, result, taskErr)
	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *MockRepository) UpdateExpression(id string, status string, result float64) (bool, error) {
	args := m.Called(id, status, result)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) CalculateFinalResult(expressionID string) (float64, error) {
	args := m.Called(expressionID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRepository) AreAllTasksCompleted(expressionID string) (bool, error) {
	args := m.Called(expressionID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) GetExpressionsByOwner(owner string) (map[string]*models.Expression, error) {
	args := m.Called(owner)
	return args.Get(0).(map[string]*models.Expression), args.Error(1)
}

func (m *MockRepository) GetExpressionByIDAndOwner(id, owner string) (*models.Expression, bool, error) {
	args := m.Called(id, owner)
	return args.Get(0).(*models.Expression), args.Bool(1), args.Error(2)
}

func (m *MockRepository) RegisterUser(user models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockRepository) FindUser(login string) (*models.User, error) {
	args := m.Called(login)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) GetTaskResult(taskID string) (float64, bool, error) {
	args := m.Called(taskID)
	return args.Get(0).(float64), args.Bool(1), args.Error(2)
}

func TestAddExpression_Success(t *testing.T) {
	mockRepo := new(MockRepository)

	mockRepo.On("AddExpression", mock.Anything).Return(nil)
	mockRepo.On("AddTask", mock.Anything).Return(nil)

	orc := service.NewOrchestrator(10, 10, 10, 10, mockRepo)

	id, err := orc.AddExpression("2 + 2", "test_user")

	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	mockRepo.AssertExpectations(t)
}
