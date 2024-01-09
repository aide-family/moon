package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/agent"
)

type AlarmService struct {
	pb.UnimplementedAlarmServer

	log *log.Helper
}

func NewAlarmService(logger log.Logger) *AlarmService {
	return &AlarmService{
		log: log.NewHelper(log.With(logger, "module", "service.alarm")),
	}
}

func (s *AlarmService) Push(ctx context.Context, req *pb.PushRequest) (*pb.PushReply, error) {
	reqBytes, _ := json.Marshal(req)
	fmt.Println("=========================")
	fmt.Println(string(reqBytes))
	fmt.Println("=========================")
	return &pb.PushReply{}, nil
}
