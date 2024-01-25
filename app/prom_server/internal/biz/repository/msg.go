package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prometheus-manager/app/prom_server/internal/biz/bo"
)

var _ MsgRepo = (*UnimplementedMsgRepo)(nil)

type (
	MsgRepo interface {
		mustEmbedUnimplemented()
		// SendAlarm 发送告警消息
		SendAlarm(ctx context.Context, hookBytes []byte, req ...*bo.AlarmMsgBo) error
	}

	UnimplementedMsgRepo struct{}
)

func (UnimplementedMsgRepo) mustEmbedUnimplemented() {}

func (UnimplementedMsgRepo) SendAlarm(_ context.Context, _ []byte, _ ...*bo.AlarmMsgBo) error {
	return status.Error(codes.Unimplemented, "method SendAlarm not implemented")
}
