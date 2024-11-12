package alarm

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/alarm"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
)

type AlarmSendService struct {
	pb.UnimplementedSendServer

	alarmSendBiz  *biz.AlarmSendBiz
	alarmGroupBiz *biz.AlarmGroupBiz
}

// NewSendService new a send service.
func NewSendService(alarmSendBiz *biz.AlarmSendBiz, alarmGroupBiz *biz.AlarmGroupBiz) *AlarmSendService {
	return &AlarmSendService{alarmSendBiz: alarmSendBiz, alarmGroupBiz: alarmGroupBiz}
}
func (s *AlarmSendService) GetAlarmSendHistory(ctx context.Context, req *pb.GetAlarmSendRequest) (*pb.GetAlarmSendReply, error) {
	sendDetail, err := s.alarmSendBiz.GetAlarmSendDetail(ctx, &bo.GetAlarmSendHistoryParams{ID: req.Id})
	if err != nil {
		return nil, err
	}

	groupDetail, _ := s.alarmGroupBiz.GetAlarmGroupDetail(ctx, sendDetail.AlarmGroupID)
	return &pb.GetAlarmSendReply{
		Detail: builder.NewParamsBuild(ctx).
			AlarmSendModuleBuilder().
			WithDoAlarmSendItem(ctx).
			ToAPI(sendDetail, groupDetail),
	}, nil
}
func (s *AlarmSendService) ListSendHistory(ctx context.Context, req *pb.ListAlarmSendRequest) (*pb.ListAlarmSendReply, error) {
	param := builder.NewParamsBuild(ctx).AlarmSendModuleBuilder().WithListAlarmSendRequest(ctx, req).ToBo()
	sendHistories, err := s.alarmSendBiz.ListAlarmSendHistories(ctx, param)
	if err != nil {
		return nil, err
	}
	return &pb.ListAlarmSendReply{
		List:       builder.NewParamsBuild(ctx).AlarmSendModuleBuilder().WithDoAlarmSendItem(ctx).ToAPIs(sendHistories),
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(param.Page),
	}, nil
}
func (s *AlarmSendService) RetrySend(ctx context.Context, req *pb.RetrySendRequest) (*pb.RetrySendReply, error) {
	err := s.alarmSendBiz.RetryAlarmSend(ctx, &bo.RetryAlarmSendParams{RequestID: req.RequestId})
	if err != nil {
		return nil, err
	}
	return &pb.RetrySendReply{}, nil
}
