package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NewOuterCaptcha creates a captcha client that calls a remote goddess (external domain).
func newGoddessCaptcha(c *config.ExternalDomainConfig) (goddessv1.CaptchaServer, func() error, error) {
	outer := c
	if outer == nil {
		return nil, nil, merr.ErrorInternalServer("external domain config is required")
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("goddess.captcha", outer.GetEndpoints(), timeout, externalNetwork(outer))

	switch externalNetwork(outer) {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewCaptchaHTTPClient(httpClient)
		return &outerCaptchaServer{cfg: outer, httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewCaptchaClient(grpcConn)
		return &outerCaptchaServer{cfg: outer, grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", externalNetwork(outer))
	}
}

type outerCaptchaServer struct {
	goddessv1.UnimplementedCaptchaServer

	cfg        *config.ExternalDomainConfig
	httpClient goddessv1.CaptchaHTTPClient
	grpcClient goddessv1.CaptchaClient
}

func (o *outerCaptchaServer) GetCaptcha(ctx context.Context, req *goddessv1.GetCaptchaRequest) (*goddessv1.GetCaptchaReply, error) {
	if o.httpClient != nil {
		return o.httpClient.GetCaptcha(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.GetCaptcha(externalContext(ctx, o.cfg), req)
}
