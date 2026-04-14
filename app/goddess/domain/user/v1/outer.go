package userv1

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	userdomain "github.com/aide-family/goddess/domain/user"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func init() {
	userdomain.RegisterUserV1Factory(config.DomainConfig_OUTER, NewOuterUser)
}

// NewOuterUser creates a user client that calls a remote goddess (OUTER driver).
func NewOuterUser(c *config.DomainConfig, driver *anypb.Any) (goddessv1.UserServer, func() error, error) {
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
	initCfg := connect.NewDefaultConfig("goddess.user", outer.GetAddress(), timeout, outer.GetProtocol().String())

	switch outer.GetProtocol().String() {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewUserHTTPClient(httpClient)
		return &outerUserServer{httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewUserClient(grpcConn)
		return &outerUserServer{grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", outer.GetProtocol().String())
	}
}

type outerUserServer struct {
	goddessv1.UnimplementedUserServer

	httpClient goddessv1.UserHTTPClient
	grpcClient goddessv1.UserClient
}

func (o *outerUserServer) GetUser(ctx context.Context, req *goddessv1.GetUserRequest) (*goddessv1.UserItem, error) {
	if pointer.IsNotNil(o.httpClient) {
		return o.httpClient.GetUser(ctx, req)
	}
	return o.grpcClient.GetUser(ctx, req)
}

func (o *outerUserServer) ListUser(ctx context.Context, req *goddessv1.ListUserRequest) (*goddessv1.ListUserReply, error) {
	if pointer.IsNotNil(o.httpClient) {
		return o.httpClient.ListUser(ctx, req)
	}
	return o.grpcClient.ListUser(ctx, req)
}

func (o *outerUserServer) SelectUser(ctx context.Context, req *goddessv1.SelectUserRequest) (*goddessv1.SelectUserReply, error) {
	if pointer.IsNotNil(o.httpClient) {
		return o.httpClient.SelectUser(ctx, req)
	}
	return o.grpcClient.SelectUser(ctx, req)
}

func (o *outerUserServer) BanUser(ctx context.Context, req *goddessv1.BanUserRequest) (*goddessv1.BanUserReply, error) {
	if pointer.IsNotNil(o.httpClient) {
		return o.httpClient.BanUser(ctx, req)
	}
	return o.grpcClient.BanUser(ctx, req)
}

func (o *outerUserServer) PermitUser(ctx context.Context, req *goddessv1.PermitUserRequest) (*goddessv1.PermitUserReply, error) {
	if pointer.IsNotNil(o.httpClient) {
		return o.httpClient.PermitUser(ctx, req)
	}
	return o.grpcClient.PermitUser(ctx, req)
}
