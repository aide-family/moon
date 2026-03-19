package webhookv1

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	webhookdomain "github.com/aide-family/rabbit/domain/webhook"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func init() {
	webhookdomain.RegisterWebhookV1Factory(config.DomainConfig_OUTER, NewOuterWebhook)
}

// NewOuterWebhook creates a webhook client that calls a remote rabbit (OUTER driver).
func NewOuterWebhook(c *config.DomainConfig, driver *anypb.Any) (apiv1.WebhookServer, func() error, error) {
	outer := &config.OuterServerConfig{}
	if pointer.IsNotNil(driver) {
		if err := anypb.UnmarshalTo(driver, outer, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal outer server config failed: %v", err)
		}
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("rabbit.webhook", outer.GetAddress(), timeout, outer.GetProtocol().String())

	switch outer.GetProtocol().String() {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewWebhookHTTPClient(httpClient)
		return &outerWebhookServer{httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewWebhookClient(grpcConn)
		return &outerWebhookServer{grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", outer.GetProtocol().String())
	}
}

type outerWebhookServer struct {
	apiv1.UnimplementedWebhookServer

	httpClient apiv1.WebhookHTTPClient
	grpcClient apiv1.WebhookClient
}

func (o *outerWebhookServer) CreateWebhook(ctx context.Context, req *apiv1.CreateWebhookRequest) (*apiv1.CreateWebhookReply, error) {
	if o.httpClient != nil {
		return o.httpClient.CreateWebhook(ctx, req)
	}
	return o.grpcClient.CreateWebhook(ctx, req)
}

func (o *outerWebhookServer) UpdateWebhook(ctx context.Context, req *apiv1.UpdateWebhookRequest) (*apiv1.UpdateWebhookReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateWebhook(ctx, req)
	}
	return o.grpcClient.UpdateWebhook(ctx, req)
}

func (o *outerWebhookServer) UpdateWebhookStatus(ctx context.Context, req *apiv1.UpdateWebhookStatusRequest) (*apiv1.UpdateWebhookStatusReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateWebhookStatus(ctx, req)
	}
	return o.grpcClient.UpdateWebhookStatus(ctx, req)
}

func (o *outerWebhookServer) DeleteWebhook(ctx context.Context, req *apiv1.DeleteWebhookRequest) (*apiv1.DeleteWebhookReply, error) {
	if o.httpClient != nil {
		return o.httpClient.DeleteWebhook(ctx, req)
	}
	return o.grpcClient.DeleteWebhook(ctx, req)
}

func (o *outerWebhookServer) GetWebhook(ctx context.Context, req *apiv1.GetWebhookRequest) (*apiv1.WebhookItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetWebhook(ctx, req)
	}
	return o.grpcClient.GetWebhook(ctx, req)
}

func (o *outerWebhookServer) ListWebhook(ctx context.Context, req *apiv1.ListWebhookRequest) (*apiv1.ListWebhookReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListWebhook(ctx, req)
	}
	return o.grpcClient.ListWebhook(ctx, req)
}

func (o *outerWebhookServer) SelectWebhook(ctx context.Context, req *apiv1.SelectWebhookRequest) (*apiv1.SelectWebhookReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SelectWebhook(ctx, req)
	}
	return o.grpcClient.SelectWebhook(ctx, req)
}
