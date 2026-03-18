package captchav1

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
	RegisterCaptchaFactoryV1(config.DomainConfig_OUTER, NewOuterCaptcha)
}

// NewOuterCaptcha creates a captcha client that calls a remote goddess (OUTER driver).
func NewOuterCaptcha(c *config.DomainConfig) (goddessv1.CaptchaServer, func() error, error) {
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
	initCfg := connect.NewDefaultConfig("goddess.captcha", outer.GetAddress(), timeout, outer.GetProtocol().String())

	switch outer.GetProtocol().String() {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewCaptchaHTTPClient(httpClient)
		return &outerCaptchaServer{httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewCaptchaClient(grpcConn)
		return &outerCaptchaServer{grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", outer.GetProtocol().String())
	}
}

type outerCaptchaServer struct {
	goddessv1.UnimplementedCaptchaServer

	httpClient goddessv1.CaptchaHTTPClient
	grpcClient goddessv1.CaptchaClient
}

func (o *outerCaptchaServer) GetCaptcha(ctx context.Context, req *goddessv1.GetCaptchaRequest) (*goddessv1.GetCaptchaReply, error) {
	if o.httpClient != nil {
		return o.httpClient.GetCaptcha(ctx, req)
	}
	return o.grpcClient.GetCaptcha(ctx, req)
}
