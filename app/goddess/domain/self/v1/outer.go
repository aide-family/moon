package selfv1

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func init() {
	RegisterSelfFactoryV1(config.DomainConfig_OUTER, NewOuterSelf)
}

// NewOuterSelf creates a self client that calls a remote goddess (OUTER driver).
func NewOuterSelf(c *config.DomainConfig) (goddessv1.SelfServer, func() error, error) {
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
	initCfg := connect.NewDefaultConfig("goddess.self", outer.GetAddress(), timeout, outer.GetProtocol().String())

	switch outer.GetProtocol().String() {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewSelfHTTPClient(httpClient)
		return &outerSelfServer{httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewSelfClient(grpcConn)
		return &outerSelfServer{grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", outer.GetProtocol().String())
	}
}

type outerSelfServer struct {
	goddessv1.UnimplementedSelfServer

	httpClient goddessv1.SelfHTTPClient
	grpcClient goddessv1.SelfClient
}

func (o *outerSelfServer) Info(ctx context.Context, req *goddessv1.InfoRequest) (*goddessv1.UserItem, error) {
	if o.httpClient != nil {
		return o.httpClient.Info(ctx, req)
	}
	return o.grpcClient.Info(ctx, req)
}

func (o *outerSelfServer) Namespaces(ctx context.Context, req *goddessv1.InfoRequest) (*goddessv1.NamespacesReply, error) {
	if o.httpClient != nil {
		return o.httpClient.Namespaces(ctx, req)
	}
	return o.grpcClient.Namespaces(ctx, req)
}

func (o *outerSelfServer) ChangeEmail(ctx context.Context, req *goddessv1.ChangeEmailRequest) (*goddessv1.ChangeReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ChangeEmail(ctx, req)
	}
	return o.grpcClient.ChangeEmail(ctx, req)
}

func (o *outerSelfServer) ChangeAvatar(ctx context.Context, req *goddessv1.ChangeAvatarRequest) (*goddessv1.ChangeReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ChangeAvatar(ctx, req)
	}
	return o.grpcClient.ChangeAvatar(ctx, req)
}

func (o *outerSelfServer) ChangeRemark(ctx context.Context, req *goddessv1.ChangeRemarkRequest) (*goddessv1.ChangeReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ChangeRemark(ctx, req)
	}
	return o.grpcClient.ChangeRemark(ctx, req)
}

func (o *outerSelfServer) RefreshToken(ctx context.Context, req *goddessv1.RefreshTokenRequest) (*goddessv1.RefreshTokenReply, error) {
	if o.httpClient != nil {
		return o.httpClient.RefreshToken(ctx, req)
	}
	return o.grpcClient.RefreshToken(ctx, req)
}
