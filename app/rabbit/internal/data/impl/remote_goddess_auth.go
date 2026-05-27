package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/oauth"
)

// NewOuterAuth creates an auth client that calls a remote goddess (external domain).
func newGoddessAuth(c *config.ExternalDomainConfig) (goddessv1.AuthServiceServer, func() error, error) {
	outer := c
	if outer == nil {
		return nil, nil, merr.ErrorInternalServer("external domain config is required")
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("goddess.auth", outer.GetEndpoints(), timeout, externalNetwork(outer))

	switch externalNetwork(outer) {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewAuthServiceHTTPClient(httpClient)
		return &outerAuthServer{cfg: outer, httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewAuthServiceClient(grpcConn)
		return &outerAuthServer{cfg: outer, grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", externalNetwork(outer))
	}
}

type outerAuthServer struct {
	goddessv1.UnimplementedAuthServiceServer

	cfg        *config.ExternalDomainConfig
	httpClient goddessv1.AuthServiceHTTPClient
	grpcClient goddessv1.AuthServiceClient
}

func (o *outerAuthServer) OAuth2Login(ctx context.Context, req *oauth.OAuth2LoginRequest) (*goddessv1.LoginReply, error) {
	if o.httpClient != nil {
		return o.httpClient.OAuth2Login(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.OAuth2Login(externalContext(ctx, o.cfg), req)
}

func (o *outerAuthServer) SendEmailLoginCode(ctx context.Context, req *goddessv1.SendEmailLoginCodeRequest) (*goddessv1.SendEmailLoginCodeReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendEmailLoginCode(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.SendEmailLoginCode(externalContext(ctx, o.cfg), req)
}

func (o *outerAuthServer) EmailLogin(ctx context.Context, req *goddessv1.EmailLoginRequest) (*goddessv1.LoginReply, error) {
	if o.httpClient != nil {
		return o.httpClient.EmailLogin(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.EmailLogin(externalContext(ctx, o.cfg), req)
}
