package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

// NewOuterSender creates a sender client that calls a remote rabbit (external domain).
func newRabbitSender(c *config.ExternalDomainConfig) (apiv1.SenderServer, func() error, error) {
	outer := c
	if outer == nil {
		return nil, nil, merr.ErrorInternalServer("external domain config is required")
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("rabbit.sender", outer.GetEndpoints(), timeout, externalNetwork(outer))

	switch externalNetwork(outer) {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewSenderHTTPClient(httpClient)
		return &outerSenderServer{cfg: outer, httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewSenderClient(grpcConn)
		return &outerSenderServer{cfg: outer, grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", externalNetwork(outer))
	}
}

type outerSenderServer struct {
	apiv1.UnimplementedSenderServer

	cfg        *config.ExternalDomainConfig
	httpClient apiv1.SenderHTTPClient
	grpcClient apiv1.SenderClient
}

func (o *outerSenderServer) SendMessage(ctx context.Context, req *apiv1.SendMessageRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendMessage(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.SendMessage(externalContext(ctx, o.cfg), req)
}

func (o *outerSenderServer) SendEmail(ctx context.Context, req *apiv1.SendEmailRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendEmail(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.SendEmail(externalContext(ctx, o.cfg), req)
}

func (o *outerSenderServer) SendEmailWithTemplate(ctx context.Context, req *apiv1.SendEmailWithTemplateRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendEmailWithTemplate(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.SendEmailWithTemplate(externalContext(ctx, o.cfg), req)
}

func (o *outerSenderServer) SendWebhook(ctx context.Context, req *apiv1.SendWebhookRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendWebhook(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.SendWebhook(externalContext(ctx, o.cfg), req)
}

func (o *outerSenderServer) SendWebhookWithTemplate(ctx context.Context, req *apiv1.SendWebhookWithTemplateRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendWebhookWithTemplate(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.SendWebhookWithTemplate(externalContext(ctx, o.cfg), req)
}
