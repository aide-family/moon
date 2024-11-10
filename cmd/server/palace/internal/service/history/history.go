package history

import (
	"context"

	historyapi "github.com/aide-family/moon/api/admin/history"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
)

// Service is a history service.
type Service struct {
	historyapi.UnimplementedHistoryServer

	alarmHistoryBiz *biz.AlarmHistoryBiz
}

// NewHistoryService new a history service.
func NewHistoryService(alarmHistoryBiz *biz.AlarmHistoryBiz) *Service {
	return &Service{alarmHistoryBiz: alarmHistoryBiz}
}

func (s *Service) GetHistory(ctx context.Context, req *historyapi.GetHistoryRequest) (*historyapi.GetHistoryReply, error) {
	param := builder.NewParamsBuild(ctx).
		AlarmHistoryModuleBuilder().
		WithGetAlarmHistoryRequest(req).
		ToBo()
	history, err := s.alarmHistoryBiz.GetAlarmHistory(ctx, param)
	if err != nil {
		return nil, err
	}
	return &historyapi.GetHistoryReply{
		AlarmHistory: builder.
			NewParamsBuild(ctx).
			AlarmHistoryModuleBuilder().
			DoAlarmHistoryItemBuilder().
			ToAPI(history),
	}, nil
}
func (s *Service) ListHistory(ctx context.Context, req *historyapi.ListHistoryRequest) (*historyapi.ListHistoryReply, error) {
	param := builder.NewParamsBuild(ctx).
		AlarmHistoryModuleBuilder().
		WithListAlarmHistoryRequest(req).
		ToBo()
	histories, err := s.alarmHistoryBiz.ListAlarmHistories(ctx, param)
	if err != nil {
		return nil, err
	}
	return &historyapi.ListHistoryReply{
		List: builder.NewParamsBuild(ctx).
			AlarmHistoryModuleBuilder().
			DoAlarmHistoryItemBuilder().
			ToAPIs(histories),
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(param.Page),
	}, nil
}
