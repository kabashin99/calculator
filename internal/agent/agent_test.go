package agent_test

import (
	"calculator_app/internal/agent"
	"calculator_app/internal/agent/mocks"
	"calculator_app/internal/pkg/models"
	pb "calculator_app/internal/proto"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"testing"
	"time"
)

//go:generate  mockgen calculator_app/internal/proto OrchestratorServiceClient > internal/agent/mocks/mock_orchestrator.go

func TestExecuteTask(t *testing.T) {
	a := &agent.Agent{}

	tests := []struct {
		name     string
		task     *models.Task
		expected float64
		wantErr  bool
	}{
		{"Addition", &models.Task{Arg1: 3, Arg2: 2, Operation: "+"}, 5, false},
		{"Subtraction", &models.Task{Arg1: 3, Arg2: 2, Operation: "-"}, 1, false},
		{"Multiplication", &models.Task{Arg1: 3, Arg2: 2, Operation: "*"}, 6, false},
		{"Division", &models.Task{Arg1: 4, Arg2: 2, Operation: "/"}, 2, false},
		{"DivisionByZero", &models.Task{Arg1: 4, Arg2: 0, Operation: "/"}, 0, true},
		{"UnknownOperation", &models.Task{Arg1: 4, Arg2: 2, Operation: "%"}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := a.ExecuteTask(tt.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteTask() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFetchTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOrchestratorServiceClient(ctrl)

	mockResp := &pb.GetTaskResponse{
		TaskId:        "1",
		Operation:     "+",
		Arg1:          1.0,
		Arg2:          2.0,
		OperationTime: 100,
		DependsOn:     []string{},
		UserLogin:     "test_user",
	}

	mockClient.EXPECT().
		GetTask(gomock.Any(), gomock.Any()).
		Return(mockResp, nil)

	a := &agent.Agent{Client: mockClient}
	task, err := a.FetchTask()
	assert.NoError(t, err)
	assert.Equal(t, "1", task.ID)
	assert.Equal(t, "+", task.Operation)
}

func TestStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOrchestratorServiceClient(ctrl)

	mockClient.EXPECT().GetTask(gomock.Any(), gomock.Any()).Return(&pb.GetTaskResponse{
		TaskId:        "test",
		Operation:     "+",
		Arg1:          1,
		Arg2:          2,
		OperationTime: 10,
		DependsOn:     nil,
	}, nil).AnyTimes()

	mockClient.EXPECT().SubmitResult(gomock.Any(), gomock.Any()).Return(&pb.SubmitResultResponse{}, nil).AnyTimes()

	testAgent := agent.NewTestAgent(mockClient, 1)

	go testAgent.Start()

	time.Sleep(300 * time.Millisecond)

	testAgent.Stop()
}

func TestSubmitWithRetry_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOrchestratorServiceClient(ctrl)

	mockClient.EXPECT().SubmitResult(gomock.Any(), gomock.Any()).Return(&pb.SubmitResultResponse{}, nil).Times(1)

	testAgent := agent.NewTestAgent(mockClient, 1)

	result := 10.0
	err := testAgent.SubmitWithRetry("task1", &result, 3, nil)

	assert.NoError(t, err)
}

func TestSubmitWithRetry_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOrchestratorServiceClient(ctrl)

	mockClient.EXPECT().SubmitResult(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("network error")).Times(3)

	testAgent := agent.NewTestAgent(mockClient, 1)

	result := 10.0
	err := testAgent.SubmitWithRetry("task1", &result, 3, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "after 3 attempts")
}

func TestSubmitResult_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOrchestratorServiceClient(ctrl)

	expectedReq := &pb.SubmitResultRequest{
		TaskId: "task123",
		Outcome: &pb.SubmitResultRequest_Result{
			Result: 42.0,
		},
	}

	mockClient.EXPECT().
		SubmitResult(gomock.Any(), expectedReq).
		Return(&pb.SubmitResultResponse{Success: true}, nil)

	testAgent := agent.NewTestAgent(mockClient, 1)

	result := 42.0
	err := testAgent.SubmitResult("task123", &result)

	assert.NoError(t, err)
}

func TestSubmitError_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOrchestratorServiceClient(ctrl)

	expectedReq := &pb.SubmitResultRequest{
		TaskId: "task123",
		Outcome: &pb.SubmitResultRequest_Error{
			Error: "division_by_zero",
		},
	}

	mockClient.EXPECT().
		SubmitResult(gomock.Any(), expectedReq).
		Return(&pb.SubmitResultResponse{Success: true}, nil)

	testAgent := agent.NewTestAgent(mockClient, 1)

	taskErr := &models.TaskError{Code: models.ErrDivisionByZero, Message: "division by zero"}

	err := testAgent.SubmitError("task123", taskErr)

	assert.NoError(t, err)
}

func TestGetDependencyResult_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOrchestratorServiceClient(ctrl)

	mockClient.EXPECT().
		GetTaskResult(gomock.Any(), &pb.GetTaskResultRequest{TaskId: "dep1"}).
		Return(&pb.GetTaskResultResponse{
			Result:     &wrapperspb.DoubleValue{Value: 7.5},
			TaskExists: true,
		}, nil)

	testAgent := agent.NewTestAgent(mockClient, 1)

	result, err := testAgent.GetDependencyResult("dep1")
	assert.NoError(t, err)
	assert.Equal(t, 7.5, result)
}
