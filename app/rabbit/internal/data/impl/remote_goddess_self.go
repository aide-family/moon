package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NewOuterSelf creates a self client that calls a remote goddess (external domain).
func newGoddessSelf(c *config.ExternalDomainConfig) (goddessv1.SelfServer, func() error, error) {
	outer := c
	if outer == nil {
		return nil, nil, merr.ErrorInternalServer("external domain config is required")
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("goddess.self", outer.GetEndpoints(), timeout, externalNetwork(outer))

	switch externalNetwork(outer) {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewSelfHTTPClient(httpClient)
		return &outerSelfServer{cfg: outer, httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewSelfClient(grpcConn)
		return &outerSelfServer{cfg: outer, grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", externalNetwork(outer))
	}
}

type outerSelfServer struct {
	goddessv1.UnimplementedSelfServer

	cfg        *config.ExternalDomainConfig
	httpClient goddessv1.SelfHTTPClient
	grpcClient goddessv1.SelfClient
}

func (o *outerSelfServer) Info(ctx context.Context, req *goddessv1.InfoRequest) (*goddessv1.UserItem, error) {
	if o.httpClient != nil {
		return o.httpClient.Info(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.Info(externalContext(ctx, o.cfg), req)
}

func (o *outerSelfServer) Namespaces(ctx context.Context, req *goddessv1.InfoRequest) (*goddessv1.NamespacesReply, error) {
	if o.httpClient != nil {
		return o.httpClient.Namespaces(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.Namespaces(externalContext(ctx, o.cfg), req)
}

func (o *outerSelfServer) ChangeEmail(ctx context.Context, req *goddessv1.ChangeEmailRequest) (*goddessv1.ChangeReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ChangeEmail(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.ChangeEmail(externalContext(ctx, o.cfg), req)
}

func (o *outerSelfServer) ChangeAvatar(ctx context.Context, req *goddessv1.ChangeAvatarRequest) (*goddessv1.ChangeReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ChangeAvatar(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.ChangeAvatar(externalContext(ctx, o.cfg), req)
}

func (o *outerSelfServer) ChangeRemark(ctx context.Context, req *goddessv1.ChangeRemarkRequest) (*goddessv1.ChangeReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ChangeRemark(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.ChangeRemark(externalContext(ctx, o.cfg), req)
}

func (o *outerSelfServer) RefreshToken(ctx context.Context, req *goddessv1.RefreshTokenRequest) (*goddessv1.RefreshTokenReply, error) {
	if o.httpClient != nil {
		return o.httpClient.RefreshToken(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.RefreshToken(externalContext(ctx, o.cfg), req)
}
