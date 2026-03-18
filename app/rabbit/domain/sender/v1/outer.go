package senderv1

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func init() {
	RegisterSenderV1Factory(config.DomainConfig_OUTER, NewOuterSender)
}

// NewOuterSender creates a sender client that calls a remote rabbit (OUTER driver).
func NewOuterSender(c *config.DomainConfig) (apiv1.SenderServer, func() error, error) {
	outer := &config.OuterServerConfig{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), outer, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal outer server config failed: %v", err)
		}
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("rabbit.sender", outer.GetAddress(), timeout, outer.GetProtocol().String())

	switch outer.GetProtocol().String() {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewSenderHTTPClient(httpClient)
		return &outerSenderServer{httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewSenderClient(grpcConn)
		return &outerSenderServer{grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", outer.GetProtocol().String())
	}
}

type outerSenderServer struct {
	apiv1.UnimplementedSenderServer

	httpClient apiv1.SenderHTTPClient
	grpcClient apiv1.SenderClient
}

func (o *outerSenderServer) SendMessage(ctx context.Context, req *apiv1.SendMessageRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendMessage(ctx, req)
	}
	return o.grpcClient.SendMessage(ctx, req)
}

func (o *outerSenderServer) SendEmail(ctx context.Context, req *apiv1.SendEmailRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendEmail(ctx, req)
	}
	return o.grpcClient.SendEmail(ctx, req)
}

func (o *outerSenderServer) SendEmailWithTemplate(ctx context.Context, req *apiv1.SendEmailWithTemplateRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendEmailWithTemplate(ctx, req)
	}
	return o.grpcClient.SendEmailWithTemplate(ctx, req)
}

func (o *outerSenderServer) SendWebhook(ctx context.Context, req *apiv1.SendWebhookRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendWebhook(ctx, req)
	}
	return o.grpcClient.SendWebhook(ctx, req)
}

func (o *outerSenderServer) SendWebhookWithTemplate(ctx context.Context, req *apiv1.SendWebhookWithTemplateRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendWebhookWithTemplate(ctx, req)
	}
	return o.grpcClient.SendWebhookWithTemplate(ctx, req)
}
