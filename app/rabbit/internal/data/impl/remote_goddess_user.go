package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NewOuterUser creates a user client that calls a remote goddess (external domain).
func newGoddessUser(c *config.ExternalDomainConfig) (goddessv1.UserServer, func() error, error) {
	outer := c
	if outer == nil {
		return nil, nil, merr.ErrorInternalServer("external domain config is required")
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("goddess.user", outer.GetEndpoints(), timeout, externalNetwork(outer))

	switch externalNetwork(outer) {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewUserHTTPClient(httpClient)
		return &outerUserServer{cfg: outer, httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewUserClient(grpcConn)
		return &outerUserServer{cfg: outer, grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", externalNetwork(outer))
	}
}

type outerUserServer struct {
	goddessv1.UnimplementedUserServer

	cfg        *config.ExternalDomainConfig
	httpClient goddessv1.UserHTTPClient
	grpcClient goddessv1.UserClient
}

func (o *outerUserServer) GetUser(ctx context.Context, req *goddessv1.GetUserRequest) (*goddessv1.UserItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetUser(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.GetUser(externalContext(ctx, o.cfg), req)
}

func (o *outerUserServer) ListUser(ctx context.Context, req *goddessv1.ListUserRequest) (*goddessv1.ListUserReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListUser(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.ListUser(externalContext(ctx, o.cfg), req)
}

func (o *outerUserServer) SelectUser(ctx context.Context, req *goddessv1.SelectUserRequest) (*goddessv1.SelectUserReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SelectUser(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.SelectUser(externalContext(ctx, o.cfg), req)
}

func (o *outerUserServer) BanUser(ctx context.Context, req *goddessv1.BanUserRequest) (*goddessv1.BanUserReply, error) {
	if o.httpClient != nil {
		return o.httpClient.BanUser(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.BanUser(externalContext(ctx, o.cfg), req)
}

func (o *outerUserServer) PermitUser(ctx context.Context, req *goddessv1.PermitUserRequest) (*goddessv1.PermitUserReply, error) {
	if o.httpClient != nil {
		return o.httpClient.PermitUser(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.PermitUser(externalContext(ctx, o.cfg), req)
}
