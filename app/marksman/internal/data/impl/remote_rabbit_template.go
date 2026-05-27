package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

// NewOuterTemplate creates a template client that calls a remote rabbit (external domain).
func newRabbitTemplate(c *config.ExternalDomainConfig) (apiv1.TemplateServer, func() error, error) {
	outer := c
	if outer == nil {
		return nil, nil, merr.ErrorInternalServer("external domain config is required")
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("rabbit.template", outer.GetEndpoints(), timeout, externalNetwork(outer))

	switch externalNetwork(outer) {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewTemplateHTTPClient(httpClient)
		return &outerTemplateServer{cfg: outer, httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewTemplateClient(grpcConn)
		return &outerTemplateServer{cfg: outer, grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", externalNetwork(outer))
	}
}

type outerTemplateServer struct {
	apiv1.UnimplementedTemplateServer

	cfg        *config.ExternalDomainConfig
	httpClient apiv1.TemplateHTTPClient
	grpcClient apiv1.TemplateClient
}

func (o *outerTemplateServer) CreateTemplate(ctx context.Context, req *apiv1.CreateTemplateRequest) (*apiv1.CreateTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.CreateTemplate(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.CreateTemplate(externalContext(ctx, o.cfg), req)
}

func (o *outerTemplateServer) UpdateTemplate(ctx context.Context, req *apiv1.UpdateTemplateRequest) (*apiv1.UpdateTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateTemplate(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.UpdateTemplate(externalContext(ctx, o.cfg), req)
}

func (o *outerTemplateServer) UpdateTemplateStatus(ctx context.Context, req *apiv1.UpdateTemplateStatusRequest) (*apiv1.UpdateTemplateStatusReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateTemplateStatus(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.UpdateTemplateStatus(externalContext(ctx, o.cfg), req)
}

func (o *outerTemplateServer) DeleteTemplate(ctx context.Context, req *apiv1.DeleteTemplateRequest) (*apiv1.DeleteTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.DeleteTemplate(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.DeleteTemplate(externalContext(ctx, o.cfg), req)
}

func (o *outerTemplateServer) GetTemplate(ctx context.Context, req *apiv1.GetTemplateRequest) (*apiv1.TemplateItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetTemplate(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.GetTemplate(externalContext(ctx, o.cfg), req)
}

func (o *outerTemplateServer) ListTemplate(ctx context.Context, req *apiv1.ListTemplateRequest) (*apiv1.ListTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListTemplate(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.ListTemplate(externalContext(ctx, o.cfg), req)
}

func (o *outerTemplateServer) SelectTemplate(ctx context.Context, req *apiv1.SelectTemplateRequest) (*apiv1.SelectTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SelectTemplate(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.SelectTemplate(externalContext(ctx, o.cfg), req)
}
