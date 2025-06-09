package rpc

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/api/common"
	rabbitcommon "github.com/aide-family/moon/pkg/api/rabbit/common"
	rabbitv1 "github.com/aide-family/moon/pkg/api/rabbit/v1"
	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/merr"
)

func NewRabbitServer(data *data.Data, logger log.Logger) repository.Rabbit {
	return &rabbitServer{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.rabbit")),
	}
}

type rabbitServer struct {
	*data.Data
	helper *log.Helper
}

// Send implements repository.Rabbit.
func (r *rabbitServer) Send() (repository.RabbitSendClient, bool) {
	server, ok := r.FirstServerConn(vobj.ServerTypeRabbit)
	if !ok {
		return nil, false
	}
	return &sendClient{server: server}, true
}

// Sync implements repository.Rabbit.
func (r *rabbitServer) Sync() (repository.RabbitSyncClient, bool) {
	server, ok := r.FirstServerConn(vobj.ServerTypeRabbit)
	if !ok {
		return nil, false
	}
	return &syncClient{server: server}, true
}

// Alert implements repository.Rabbit.
func (r *rabbitServer) Alert() (repository.RabbitAlertClient, bool) {
	server, ok := r.FirstServerConn(vobj.ServerTypeRabbit)
	if !ok {
		return nil, false
	}
	return &alertClient{server: server}, true
}

type sendClient struct {
	server *bo.Server
}

func (s *sendClient) Email(ctx context.Context, in *rabbitv1.SendEmailRequest) (*rabbitcommon.EmptyReply, error) {
	switch s.server.Config.Server.GetNetwork() {
	case config.Network_GRPC:
		return rabbitv1.NewSendClient(s.server.Conn).Email(ctx, in)
	case config.Network_HTTP:
		return rabbitv1.NewSendHTTPClient(s.server.Client).Email(ctx, in)
	default:
		return nil, merr.ErrorInternalServer("network is not supported")
	}
}

func (s *sendClient) Sms(ctx context.Context, in *rabbitv1.SendSmsRequest) (*rabbitcommon.EmptyReply, error) {
	switch s.server.Config.Server.GetNetwork() {
	case config.Network_GRPC:
		return rabbitv1.NewSendClient(s.server.Conn).Sms(ctx, in)
	case config.Network_HTTP:
		return rabbitv1.NewSendHTTPClient(s.server.Client).Sms(ctx, in)
	default:
		return nil, merr.ErrorInternalServer("network is not supported")
	}
}

func (s *sendClient) Hook(ctx context.Context, in *rabbitv1.SendHookRequest) (*rabbitcommon.EmptyReply, error) {
	switch s.server.Config.Server.GetNetwork() {
	case config.Network_GRPC:
		return rabbitv1.NewSendClient(s.server.Conn).Hook(ctx, in)
	case config.Network_HTTP:
		return rabbitv1.NewSendHTTPClient(s.server.Client).Hook(ctx, in)
	default:
		return nil, merr.ErrorInternalServer("network is not supported")
	}
}

type syncClient struct {
	server *bo.Server
}

// Hook implements repository.SyncClient.
func (s *syncClient) Hook(ctx context.Context, in *rabbitv1.SyncHookRequest) (*rabbitcommon.EmptyReply, error) {
	switch s.server.Config.Server.GetNetwork() {
	case config.Network_GRPC:
		return rabbitv1.NewSyncClient(s.server.Conn).Hook(ctx, in)
	case config.Network_HTTP:
		return rabbitv1.NewSyncHTTPClient(s.server.Client).Hook(ctx, in)
	default:
		return nil, merr.ErrorInternalServer("network is not supported")
	}
}

// NoticeGroup implements repository.SyncClient.
func (s *syncClient) NoticeGroup(ctx context.Context, in *rabbitv1.SyncNoticeGroupRequest) (*rabbitcommon.EmptyReply, error) {
	switch s.server.Config.Server.GetNetwork() {
	case config.Network_GRPC:
		return rabbitv1.NewSyncClient(s.server.Conn).NoticeGroup(ctx, in)
	case config.Network_HTTP:
		return rabbitv1.NewSyncHTTPClient(s.server.Client).NoticeGroup(ctx, in)
	default:
		return nil, merr.ErrorInternalServer("network is not supported")
	}
}

// NoticeUser implements repository.SyncClient.
func (s *syncClient) NoticeUser(ctx context.Context, in *rabbitv1.SyncNoticeUserRequest) (*rabbitcommon.EmptyReply, error) {
	switch s.server.Config.Server.GetNetwork() {
	case config.Network_GRPC:
		return rabbitv1.NewSyncClient(s.server.Conn).NoticeUser(ctx, in)
	case config.Network_HTTP:
		return rabbitv1.NewSyncHTTPClient(s.server.Client).NoticeUser(ctx, in)
	default:
		return nil, merr.ErrorInternalServer("network is not supported")
	}
}

// Remove implements repository.SyncClient.
func (s *syncClient) Remove(ctx context.Context, in *rabbitv1.RemoveRequest) (*rabbitcommon.EmptyReply, error) {
	switch s.server.Config.Server.GetNetwork() {
	case config.Network_GRPC:
		return rabbitv1.NewSyncClient(s.server.Conn).Remove(ctx, in)
	case config.Network_HTTP:
		return rabbitv1.NewSyncHTTPClient(s.server.Client).Remove(ctx, in)
	default:
		return nil, merr.ErrorInternalServer("network is not supported")
	}
}
func (s *syncClient) Sms(ctx context.Context, in *rabbitv1.SyncSmsRequest) (*rabbitcommon.EmptyReply, error) {
	switch s.server.Config.Server.GetNetwork() {
	case config.Network_GRPC:
		return rabbitv1.NewSyncClient(s.server.Conn).Sms(ctx, in)
	case config.Network_HTTP:
		return rabbitv1.NewSyncHTTPClient(s.server.Client).Sms(ctx, in)
	default:
		return nil, merr.ErrorInternalServer("network is not supported")
	}
}

// Email implements repository.SyncClient.
func (s *syncClient) Email(ctx context.Context, in *rabbitv1.SyncEmailRequest) (*rabbitcommon.EmptyReply, error) {
	switch s.server.Config.Server.GetNetwork() {
	case config.Network_GRPC:
		return rabbitv1.NewSyncClient(s.server.Conn).Email(ctx, in)
	case config.Network_HTTP:
		return rabbitv1.NewSyncHTTPClient(s.server.Client).Email(ctx, in)
	default:
		return nil, merr.ErrorInternalServer("network is not supported")
	}
}

type alertClient struct {
	server *bo.Server
}

// SendAlert implements repository.AlertClient.
func (a *alertClient) SendAlert(ctx context.Context, in *common.AlertsItem) (*rabbitcommon.EmptyReply, error) {
	switch a.server.Config.Server.GetNetwork() {
	case config.Network_GRPC:
		return rabbitv1.NewAlertClient(a.server.Conn).SendAlert(ctx, in)
	case config.Network_HTTP:
		return rabbitv1.NewAlertHTTPClient(a.server.Client).SendAlert(ctx, in)
	default:
		return nil, merr.ErrorInternalServer("network is not supported")
	}
}
