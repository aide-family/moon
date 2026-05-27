package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

// NewOuterWebhook creates a webhook client that calls a remote rabbit (external domain).
func newRabbitWebhook(c *config.ExternalDomainConfig) (apiv1.WebhookServer, func() error, error) {
	outer := c
	if outer == nil {
		return nil, nil, merr.ErrorInternalServer("external domain config is required")
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("rabbit.webhook", outer.GetEndpoints(), timeout, externalNetwork(outer))

	switch externalNetwork(outer) {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewWebhookHTTPClient(httpClient)
		return &outerWebhookServer{cfg: outer, httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewWebhookClient(grpcConn)
		return &outerWebhookServer{cfg: outer, grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", externalNetwork(outer))
	}
}

type outerWebhookServer struct {
	apiv1.UnimplementedWebhookServer

	cfg        *config.ExternalDomainConfig
	httpClient apiv1.WebhookHTTPClient
	grpcClient apiv1.WebhookClient
}

func (o *outerWebhookServer) CreateWebhook(ctx context.Context, req *apiv1.CreateWebhookRequest) (*apiv1.CreateWebhookReply, error) {
	if o.httpClient != nil {
		return o.httpClient.CreateWebhook(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.CreateWebhook(externalContext(ctx, o.cfg), req)
}

func (o *outerWebhookServer) UpdateWebhook(ctx context.Context, req *apiv1.UpdateWebhookRequest) (*apiv1.UpdateWebhookReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateWebhook(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.UpdateWebhook(externalContext(ctx, o.cfg), req)
}

func (o *outerWebhookServer) UpdateWebhookStatus(ctx context.Context, req *apiv1.UpdateWebhookStatusRequest) (*apiv1.UpdateWebhookStatusReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateWebhookStatus(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.UpdateWebhookStatus(externalContext(ctx, o.cfg), req)
}

func (o *outerWebhookServer) DeleteWebhook(ctx context.Context, req *apiv1.DeleteWebhookRequest) (*apiv1.DeleteWebhookReply, error) {
	if o.httpClient != nil {
		return o.httpClient.DeleteWebhook(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.DeleteWebhook(externalContext(ctx, o.cfg), req)
}

func (o *outerWebhookServer) GetWebhook(ctx context.Context, req *apiv1.GetWebhookRequest) (*apiv1.WebhookItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetWebhook(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.GetWebhook(externalContext(ctx, o.cfg), req)
}

func (o *outerWebhookServer) ListWebhook(ctx context.Context, req *apiv1.ListWebhookRequest) (*apiv1.ListWebhookReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListWebhook(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.ListWebhook(externalContext(ctx, o.cfg), req)
}

func (o *outerWebhookServer) SelectWebhook(ctx context.Context, req *apiv1.SelectWebhookRequest) (*apiv1.SelectWebhookReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SelectWebhook(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.SelectWebhook(externalContext(ctx, o.cfg), req)
}
