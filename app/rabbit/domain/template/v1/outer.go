package templatev1

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	templatedomain "github.com/aide-family/rabbit/domain/template"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func init() {
	templatedomain.RegisterTemplateV1Factory(config.DomainConfig_OUTER, NewOuterTemplate)
}

// NewOuterTemplate creates a template client that calls a remote rabbit (OUTER driver).
func NewOuterTemplate(c *config.DomainConfig, driver *anypb.Any) (apiv1.TemplateServer, func() error, error) {
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
	initCfg := connect.NewDefaultConfig("rabbit.template", outer.GetAddress(), timeout, outer.GetProtocol().String())

	switch outer.GetProtocol().String() {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewTemplateHTTPClient(httpClient)
		return &outerTemplateServer{httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewTemplateClient(grpcConn)
		return &outerTemplateServer{grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", outer.GetProtocol().String())
	}
}

type outerTemplateServer struct {
	apiv1.UnimplementedTemplateServer

	httpClient apiv1.TemplateHTTPClient
	grpcClient apiv1.TemplateClient
}

func (o *outerTemplateServer) CreateTemplate(ctx context.Context, req *apiv1.CreateTemplateRequest) (*apiv1.CreateTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.CreateTemplate(ctx, req)
	}
	return o.grpcClient.CreateTemplate(ctx, req)
}

func (o *outerTemplateServer) UpdateTemplate(ctx context.Context, req *apiv1.UpdateTemplateRequest) (*apiv1.UpdateTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateTemplate(ctx, req)
	}
	return o.grpcClient.UpdateTemplate(ctx, req)
}

func (o *outerTemplateServer) UpdateTemplateStatus(ctx context.Context, req *apiv1.UpdateTemplateStatusRequest) (*apiv1.UpdateTemplateStatusReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateTemplateStatus(ctx, req)
	}
	return o.grpcClient.UpdateTemplateStatus(ctx, req)
}

func (o *outerTemplateServer) DeleteTemplate(ctx context.Context, req *apiv1.DeleteTemplateRequest) (*apiv1.DeleteTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.DeleteTemplate(ctx, req)
	}
	return o.grpcClient.DeleteTemplate(ctx, req)
}

func (o *outerTemplateServer) GetTemplate(ctx context.Context, req *apiv1.GetTemplateRequest) (*apiv1.TemplateItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetTemplate(ctx, req)
	}
	return o.grpcClient.GetTemplate(ctx, req)
}

func (o *outerTemplateServer) ListTemplate(ctx context.Context, req *apiv1.ListTemplateRequest) (*apiv1.ListTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListTemplate(ctx, req)
	}
	return o.grpcClient.ListTemplate(ctx, req)
}

func (o *outerTemplateServer) SelectTemplate(ctx context.Context, req *apiv1.SelectTemplateRequest) (*apiv1.SelectTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SelectTemplate(ctx, req)
	}
	return o.grpcClient.SelectTemplate(ctx, req)
}
