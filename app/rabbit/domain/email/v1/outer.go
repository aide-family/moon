package emailv1

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	emaildomain "github.com/aide-family/rabbit/domain/email"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func init() {
	emaildomain.RegisterEmailV1Factory(config.DomainConfig_OUTER, NewOuterEmail)
}

// NewOuterEmail creates an email client that calls a remote rabbit (OUTER driver).
func NewOuterEmail(c *config.DomainConfig, driver *anypb.Any) (apiv1.EmailServer, func() error, error) {
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
	initCfg := connect.NewDefaultConfig("rabbit.email", outer.GetAddress(), timeout, outer.GetProtocol().String())

	switch outer.GetProtocol().String() {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewEmailHTTPClient(httpClient)
		return &outerEmailServer{httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewEmailClient(grpcConn)
		return &outerEmailServer{grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", outer.GetProtocol().String())
	}
}

type outerEmailServer struct {
	apiv1.UnimplementedEmailServer

	httpClient apiv1.EmailHTTPClient
	grpcClient apiv1.EmailClient
}

func (o *outerEmailServer) CreateEmailConfig(ctx context.Context, req *apiv1.CreateEmailConfigRequest) (*apiv1.CreateEmailConfigReply, error) {
	if o.httpClient != nil {
		return o.httpClient.CreateEmailConfig(ctx, req)
	}
	return o.grpcClient.CreateEmailConfig(ctx, req)
}

func (o *outerEmailServer) UpdateEmailConfig(ctx context.Context, req *apiv1.UpdateEmailConfigRequest) (*apiv1.UpdateEmailConfigReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateEmailConfig(ctx, req)
	}
	return o.grpcClient.UpdateEmailConfig(ctx, req)
}

func (o *outerEmailServer) UpdateEmailConfigStatus(ctx context.Context, req *apiv1.UpdateEmailConfigStatusRequest) (*apiv1.UpdateEmailConfigStatusReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateEmailConfigStatus(ctx, req)
	}
	return o.grpcClient.UpdateEmailConfigStatus(ctx, req)
}

func (o *outerEmailServer) DeleteEmailConfig(ctx context.Context, req *apiv1.DeleteEmailConfigRequest) (*apiv1.DeleteEmailConfigReply, error) {
	if o.httpClient != nil {
		return o.httpClient.DeleteEmailConfig(ctx, req)
	}
	return o.grpcClient.DeleteEmailConfig(ctx, req)
}

func (o *outerEmailServer) GetEmailConfig(ctx context.Context, req *apiv1.GetEmailConfigRequest) (*apiv1.EmailConfigItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetEmailConfig(ctx, req)
	}
	return o.grpcClient.GetEmailConfig(ctx, req)
}

func (o *outerEmailServer) ListEmailConfig(ctx context.Context, req *apiv1.ListEmailConfigRequest) (*apiv1.ListEmailConfigReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListEmailConfig(ctx, req)
	}
	return o.grpcClient.ListEmailConfig(ctx, req)
}

func (o *outerEmailServer) SelectEmailConfig(ctx context.Context, req *apiv1.SelectEmailConfigRequest) (*apiv1.SelectEmailConfigReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SelectEmailConfig(ctx, req)
	}
	return o.grpcClient.SelectEmailConfig(ctx, req)
}

