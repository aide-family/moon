package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
)

var _ MsgRepo = (*UnimplementedMsgRepo)(nil)

type (
	MsgRepo interface {
		mustEmbedUnimplemented()
		// SendAlarm 发送告警消息
		SendAlarm(ctx context.Context, req ...*bo.AlarmMsgBo) error

		SendAlarmToMember(ctx context.Context, members []*bo.NotifyMemberBO, memberTemplateMap map[vobj.NotifyType]string, alarmInfo *bo.AlertBo) error
	}

	UnimplementedMsgRepo struct{}
)

func (UnimplementedMsgRepo) SendAlarmToMember(_ context.Context, _ []*bo.NotifyMemberBO, _ map[vobj.NotifyType]string, _ *bo.AlertBo) error {
	return status.Error(codes.Unimplemented, "method SendAlarmToMember not implemented")
}

func (UnimplementedMsgRepo) mustEmbedUnimplemented() {}

func (UnimplementedMsgRepo) SendAlarm(_ context.Context, _ ...*bo.AlarmMsgBo) error {
	return status.Error(codes.Unimplemented, "method SendAlarm not implemented")
}
