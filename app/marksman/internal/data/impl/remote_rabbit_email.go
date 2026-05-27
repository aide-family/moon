package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

// NewOuterEmail creates an email client that calls a remote rabbit (external domain).
func newRabbitEmail(c *config.ExternalDomainConfig) (apiv1.EmailServer, func() error, error) {
	outer := c
	if outer == nil {
		return nil, nil, merr.ErrorInternalServer("external domain config is required")
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("rabbit.email", outer.GetEndpoints(), timeout, externalNetwork(outer))

	switch externalNetwork(outer) {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewEmailHTTPClient(httpClient)
		return &outerEmailServer{cfg: outer, httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewEmailClient(grpcConn)
		return &outerEmailServer{cfg: outer, grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", externalNetwork(outer))
	}
}

type outerEmailServer struct {
	apiv1.UnimplementedEmailServer

	cfg        *config.ExternalDomainConfig
	httpClient apiv1.EmailHTTPClient
	grpcClient apiv1.EmailClient
}

func (o *outerEmailServer) CreateEmailConfig(ctx context.Context, req *apiv1.CreateEmailConfigRequest) (*apiv1.CreateEmailConfigReply, error) {
	if o.httpClient != nil {
		return o.httpClient.CreateEmailConfig(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.CreateEmailConfig(externalContext(ctx, o.cfg), req)
}

func (o *outerEmailServer) UpdateEmailConfig(ctx context.Context, req *apiv1.UpdateEmailConfigRequest) (*apiv1.UpdateEmailConfigReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateEmailConfig(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.UpdateEmailConfig(externalContext(ctx, o.cfg), req)
}

func (o *outerEmailServer) UpdateEmailConfigStatus(ctx context.Context, req *apiv1.UpdateEmailConfigStatusRequest) (*apiv1.UpdateEmailConfigStatusReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateEmailConfigStatus(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.UpdateEmailConfigStatus(externalContext(ctx, o.cfg), req)
}

func (o *outerEmailServer) DeleteEmailConfig(ctx context.Context, req *apiv1.DeleteEmailConfigRequest) (*apiv1.DeleteEmailConfigReply, error) {
	if o.httpClient != nil {
		return o.httpClient.DeleteEmailConfig(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.DeleteEmailConfig(externalContext(ctx, o.cfg), req)
}

func (o *outerEmailServer) GetEmailConfig(ctx context.Context, req *apiv1.GetEmailConfigRequest) (*apiv1.EmailConfigItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetEmailConfig(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.GetEmailConfig(externalContext(ctx, o.cfg), req)
}

func (o *outerEmailServer) ListEmailConfig(ctx context.Context, req *apiv1.ListEmailConfigRequest) (*apiv1.ListEmailConfigReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListEmailConfig(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.ListEmailConfig(externalContext(ctx, o.cfg), req)
}

func (o *outerEmailServer) SelectEmailConfig(ctx context.Context, req *apiv1.SelectEmailConfigRequest) (*apiv1.SelectEmailConfigReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SelectEmailConfig(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.SelectEmailConfig(externalContext(ctx, o.cfg), req)
}
