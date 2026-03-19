package authv1

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	authdomain "github.com/aide-family/goddess/domain/auth"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/oauth"
)

func init() {
	authdomain.RegisterAuthV1Factory(config.DomainConfig_OUTER, NewOuterAuth)
}

// NewOuterAuth creates an auth client that calls a remote goddess (OUTER driver).
func NewOuterAuth(c *config.DomainConfig) (goddessv1.AuthServiceServer, func() error, error) {
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
	initCfg := connect.NewDefaultConfig("goddess.auth", outer.GetAddress(), timeout, outer.GetProtocol().String())

	switch outer.GetProtocol().String() {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewAuthServiceHTTPClient(httpClient)
		return &outerAuthServer{httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewAuthServiceClient(grpcConn)
		return &outerAuthServer{grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", outer.GetProtocol().String())
	}
}

type outerAuthServer struct {
	goddessv1.UnimplementedAuthServiceServer

	httpClient goddessv1.AuthServiceHTTPClient
	grpcClient goddessv1.AuthServiceClient
}

func (o *outerAuthServer) OAuth2Login(ctx context.Context, req *oauth.OAuth2LoginRequest) (*goddessv1.LoginReply, error) {
	if o.httpClient != nil {
		return o.httpClient.OAuth2Login(ctx, req)
	}
	return o.grpcClient.OAuth2Login(ctx, req)
}

func (o *outerAuthServer) SendEmailLoginCode(ctx context.Context, req *goddessv1.SendEmailLoginCodeRequest) (*goddessv1.SendEmailLoginCodeReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendEmailLoginCode(ctx, req)
	}
	return o.grpcClient.SendEmailLoginCode(ctx, req)
}

func (o *outerAuthServer) EmailLogin(ctx context.Context, req *goddessv1.EmailLoginRequest) (*goddessv1.LoginReply, error) {
	if o.httpClient != nil {
		return o.httpClient.EmailLogin(ctx, req)
	}
	return o.grpcClient.EmailLogin(ctx, req)
}
