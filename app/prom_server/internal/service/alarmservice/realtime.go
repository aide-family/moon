package alarmservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/alarm/realtime"
	"prometheus-manager/app/prom_server/internal/biz"
)

type RealtimeService struct {
	pb.UnimplementedRealtimeServer

	log           *log.Helper
	alarmRealtime *biz.AlarmRealtime
}

// NewRealtimeService 实时告警服务
func NewRealtimeService(alarmRealtime *biz.AlarmRealtime, logger log.Logger) *RealtimeService {
	return &RealtimeService{
		log:           log.NewHelper(log.With(logger, "module", "service.alarm.realtime")),
		alarmRealtime: alarmRealtime,
	}
}

// GetRealtime 实时告警详情
func (s *RealtimeService) GetRealtime(ctx context.Context, req *pb.GetRealtimeRequest) (*pb.GetRealtimeReply, error) {
	return &pb.GetRealtimeReply{}, nil
}

// ListRealtime 实时告警列表
func (s *RealtimeService) ListRealtime(ctx context.Context, req *pb.ListRealtimeRequest) (*pb.ListRealtimeReply, error) {
	return &pb.ListRealtimeReply{}, nil
}

// Intervene 告警干预
func (s *RealtimeService) Intervene(ctx context.Context, req *pb.InterveneRequest) (*pb.InterveneReply, error) {
	return &pb.InterveneReply{}, nil
}

// Upgrade 告警升级
func (s *RealtimeService) Upgrade(ctx context.Context, req *pb.UpgradeRequest) (*pb.UpgradeReply, error) {
	return &pb.UpgradeReply{}, nil
}

// Suppress 告警抑制
func (s *RealtimeService) Suppress(ctx context.Context, req *pb.SuppressRequest) (*pb.SuppressReply, error) {
	return &pb.SuppressReply{}, nil
}
