package realtime

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/realtime"
)

// AlarmService 实时告警数据服务
type AlarmService struct {
	pb.UnimplementedAlarmServer
}

// NewAlarmService 实时告警数据服务
func NewAlarmService() *AlarmService {
	return &AlarmService{}
}

// GetAlarm 获取实时告警数据
func (s *AlarmService) GetAlarm(ctx context.Context, req *pb.GetAlarmRequest) (*pb.GetAlarmReply, error) {
	return &pb.GetAlarmReply{}, nil
}

// ListAlarm 获取实时告警数据列表
func (s *AlarmService) ListAlarm(ctx context.Context, req *pb.ListAlarmRequest) (*pb.ListAlarmReply, error) {
	return &pb.ListAlarmReply{}, nil
}
