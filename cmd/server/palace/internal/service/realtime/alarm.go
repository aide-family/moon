package realtime

import (
	"context"

	realtimeapi "github.com/aide-family/moon/api/admin/realtime"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
)

// AlarmService 实时告警数据服务
type AlarmService struct {
	realtimeapi.UnimplementedAlarmServer

	alarmBiz *biz.AlarmBiz
}

// NewAlarmService 实时告警数据服务
func NewAlarmService(alarmBiz *biz.AlarmBiz) *AlarmService {
	return &AlarmService{
		alarmBiz: alarmBiz,
	}
}

// GetAlarm 获取实时告警数据
func (s *AlarmService) GetAlarm(ctx context.Context, req *realtimeapi.GetAlarmRequest) (*realtimeapi.GetAlarmReply, error) {
	params := build.NewBuilder().RealTimeAlarmModule().WithAPIGetAlarmRequest(req).ToBo()
	realtimeAlarmDetail, err := s.alarmBiz.GetRealTimeAlarm(ctx, params)
	if err != nil {
		return nil, err
	}
	return &realtimeapi.GetAlarmReply{
		Detail: build.NewBuilder().WithContext(ctx).RealTimeAlarmModule().
			WithDoRealtimeAlarm(realtimeAlarmDetail).
			ToAPI(),
	}, nil
}

// ListAlarm 获取实时告警数据列表
func (s *AlarmService) ListAlarm(ctx context.Context, req *realtimeapi.ListAlarmRequest) (*realtimeapi.ListAlarmReply, error) {
	params := build.NewBuilder().RealTimeAlarmModule().WithAPIListAlarmRequest(req).ToBo()
	list, err := s.alarmBiz.ListRealTimeAlarms(ctx, params)
	if err != nil {
		return nil, err
	}
	return &realtimeapi.ListAlarmReply{
		List: build.NewBuilder().WithContext(ctx).
			RealTimeAlarmModule().
			WithDostRealtimeAlarm(list).
			ToAPIs(),
		Pagination: build.NewPageBuilder(params.Pagination).ToAPI(),
	}, nil
}
