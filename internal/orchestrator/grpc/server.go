package grpc

import (
	"calculator_app/internal/orchestrator/service"
	pb "calculator_app/internal/proto"
	"context"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"log"
)

type OrchestratorGRPCServer struct {
	pb.UnimplementedOrchestratorServiceServer
	orc *service.Orchestrator
}

func NewOrchestratorGRPCServer(orc *service.Orchestrator) *OrchestratorGRPCServer {
	return &OrchestratorGRPCServer{orc: orc}
}

func (s *OrchestratorGRPCServer) GetTask(ctx context.Context, _ *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	task, exists, err := s.orc.GetTask()
	if err != nil || !exists {
		//log.Printf("grps server FALSE Task %t или ошибка : %s", exists, err)
		return nil, err
	}

	resp := &pb.GetTaskResponse{
		TaskId:        task.ID,
		Operation:     task.Operation,
		Arg1:          task.Arg1,
		Arg2:          task.Arg2,
		OperationTime: int32(task.OperationTime),
		DependsOn:     task.DependsOn,
		UserLogin:     task.UserLogin,
	}

	log.Printf("grpc сервер отдает таску %+v", task)
	return resp, nil
}

func (s *OrchestratorGRPCServer) SubmitResult(ctx context.Context, req *pb.SubmitResultRequest) (*pb.SubmitResultResponse, error) {
	success, err := s.orc.SubmitResult(req.TaskId, float64(req.Result))
	if err != nil {
		return nil, err
	}

	return &pb.SubmitResultResponse{Success: success}, nil
}

func (s *OrchestratorGRPCServer) GetTaskResult(ctx context.Context, req *pb.GetTaskResultRequest) (*pb.GetTaskResultResponse, error) {
	result, exists, err := s.orc.GetTaskResult(req.TaskId)
	if err != nil {
		return nil, err
	}
	var resultProto *wrapperspb.DoubleValue
	if exists {
		resultProto = wrapperspb.Double(result)
	}

	return &pb.GetTaskResultResponse{
		Result:     resultProto,
		TaskExists: exists,
	}, nil
}
