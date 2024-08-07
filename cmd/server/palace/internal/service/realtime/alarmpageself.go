package realtime

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/realtime"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/helper/middleware"
)

// AlarmPageSelfService is a service that implements the AlarmPageSelfServer.
type AlarmPageSelfService struct {
	pb.UnimplementedAlarmPageSelfServer

	alarmPageBiz *biz.AlarmPageBiz
}

// NewAlarmPageSelfService creates a new AlarmPageSelfService.
func NewAlarmPageSelfService(alarmPageBiz *biz.AlarmPageBiz) *AlarmPageSelfService {
	return &AlarmPageSelfService{
		alarmPageBiz: alarmPageBiz,
	}
}

// UpdateAlarmPage implements AlarmPageSelfServer.
func (s *AlarmPageSelfService) UpdateAlarmPage(ctx context.Context, req *pb.UpdateAlarmPageRequest) (*pb.UpdateAlarmPageReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	if err := s.alarmPageBiz.UpdateAlarmPage(ctx, claims.GetUser(), req.GetAlarmPageIds()); err != nil {
		return nil, err
	}
	return &pb.UpdateAlarmPageReply{}, nil
}

// ListAlarmPage implements AlarmPageSelfServer.
func (s *AlarmPageSelfService) ListAlarmPage(ctx context.Context, _ *pb.ListAlarmPageRequest) (*pb.ListAlarmPageReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	alarmPageList, err := s.alarmPageBiz.ListAlarmPage(ctx, claims.GetUser())
	if err != nil {
		return nil, err
	}
	return &pb.ListAlarmPageReply{
		List: build.NewBuilder().WithContext(ctx).AlarmPageModule().WithAlarmPages(alarmPageList).ToAPIs(),
	}, nil
}
