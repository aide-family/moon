package alarm

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/alarm"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
)

// SendService 告警发送
type SendService struct {
	pb.UnimplementedSendServer

	alarmSendBiz  *biz.AlarmSendBiz
	alarmGroupBiz *biz.AlarmGroupBiz
}

// NewSendService 创建告警发送服务
func NewSendService(alarmSendBiz *biz.AlarmSendBiz, alarmGroupBiz *biz.AlarmGroupBiz) *SendService {
	return &SendService{alarmSendBiz: alarmSendBiz, alarmGroupBiz: alarmGroupBiz}
}

// GetAlarmSendHistory 获取告警发送历史
func (s *SendService) GetAlarmSendHistory(ctx context.Context, req *pb.GetAlarmSendRequest) (*pb.GetAlarmSendReply, error) {
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

// ListSendHistory 获取告警发送历史列表
func (s *SendService) ListSendHistory(ctx context.Context, req *pb.ListAlarmSendRequest) (*pb.ListAlarmSendReply, error) {
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

// RetrySend 重试告警发送
func (s *SendService) RetrySend(ctx context.Context, req *pb.RetrySendRequest) (*pb.RetrySendReply, error) {
	err := s.alarmSendBiz.RetryAlarmSend(ctx, &bo.RetryAlarmSendParams{RequestID: req.RequestId})
	if err != nil {
		return nil, err
	}
	return &pb.RetrySendReply{}, nil
}
