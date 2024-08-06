package realtime

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/realtime"
)

// AlarmPageSelfService is a service that implements the AlarmPageSelfServer.
type AlarmPageSelfService struct {
	pb.UnimplementedAlarmPageSelfServer
}

// NewAlarmPageSelfService creates a new AlarmPageSelfService.
func NewAlarmPageSelfService() *AlarmPageSelfService {
	return &AlarmPageSelfService{}
}

// UpdateAlarmPage implements AlarmPageSelfServer.
func (s *AlarmPageSelfService) UpdateAlarmPage(ctx context.Context, req *pb.UpdateAlarmPageRequest) (*pb.UpdateAlarmPageReply, error) {
	return &pb.UpdateAlarmPageReply{}, nil
}

// ListAlarmPage implements AlarmPageSelfServer.
func (s *AlarmPageSelfService) ListAlarmPage(ctx context.Context, req *pb.ListAlarmPageRequest) (*pb.ListAlarmPageReply, error) {
	return &pb.ListAlarmPageReply{}, nil
}
