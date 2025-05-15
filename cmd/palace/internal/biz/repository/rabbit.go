package repository

import (
	"context"

	"github.com/moon-monitor/moon/pkg/api/common"
	rabbitcommon "github.com/moon-monitor/moon/pkg/api/rabbit/common"
	rabbitv1 "github.com/moon-monitor/moon/pkg/api/rabbit/v1"
)

type Rabbit interface {
	Send() (SendClient, bool)
	Sync() (SyncClient, bool)
	Alert() (AlertClient, bool)
}

type SendClient interface {
	Email(ctx context.Context, in *rabbitv1.SendEmailRequest) (*rabbitcommon.EmptyReply, error)
	Sms(ctx context.Context, in *rabbitv1.SendSmsRequest) (*rabbitcommon.EmptyReply, error)
	Hook(ctx context.Context, in *rabbitv1.SendHookRequest) (*rabbitcommon.EmptyReply, error)
}

type SyncClient interface {
	Sms(ctx context.Context, in *rabbitv1.SyncSmsRequest) (*rabbitcommon.EmptyReply, error)
	Email(ctx context.Context, in *rabbitv1.SyncEmailRequest) (*rabbitcommon.EmptyReply, error)
	Hook(ctx context.Context, in *rabbitv1.SyncHookRequest) (*rabbitcommon.EmptyReply, error)
	NoticeGroup(ctx context.Context, in *rabbitv1.SyncNoticeGroupRequest) (*rabbitcommon.EmptyReply, error)
	NoticeUser(ctx context.Context, in *rabbitv1.SyncNoticeUserRequest) (*rabbitcommon.EmptyReply, error)
	Remove(ctx context.Context, in *rabbitv1.RemoveRequest) (*rabbitcommon.EmptyReply, error)
}

type AlertClient interface {
	SendAlert(ctx context.Context, in *common.AlertsItem) (*rabbitcommon.EmptyReply, error)
}
