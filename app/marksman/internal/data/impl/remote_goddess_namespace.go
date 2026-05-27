package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NewOuterNamespace creates a namespace client that calls a remote goddess (external domain).
func newGoddessNamespace(c *config.ExternalDomainConfig) (goddessv1.NamespaceServer, func() error, error) {
	outer := c
	if outer == nil {
		return nil, nil, merr.ErrorInternalServer("external domain config is required")
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("goddess.namespace", outer.GetEndpoints(), timeout, externalNetwork(outer))

	switch externalNetwork(outer) {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewNamespaceHTTPClient(httpClient)
		return &outerNamespaceServer{cfg: outer, httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewNamespaceClient(grpcConn)
		return &outerNamespaceServer{cfg: outer, grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", externalNetwork(outer))
	}
}

type outerNamespaceServer struct {
	goddessv1.UnimplementedNamespaceServer

	cfg        *config.ExternalDomainConfig
	httpClient goddessv1.NamespaceHTTPClient
	grpcClient goddessv1.NamespaceClient
}

func (o *outerNamespaceServer) CreateNamespace(ctx context.Context, req *goddessv1.CreateNamespaceRequest) (*goddessv1.CreateNamespaceReply, error) {
	if o.httpClient != nil {
		return o.httpClient.CreateNamespace(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.CreateNamespace(externalContext(ctx, o.cfg), req)
}

func (o *outerNamespaceServer) UpdateNamespace(ctx context.Context, req *goddessv1.UpdateNamespaceRequest) (*goddessv1.UpdateNamespaceReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateNamespace(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.UpdateNamespace(externalContext(ctx, o.cfg), req)
}

func (o *outerNamespaceServer) UpdateNamespaceStatus(ctx context.Context, req *goddessv1.UpdateNamespaceStatusRequest) (*goddessv1.UpdateNamespaceStatusReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateNamespaceStatus(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.UpdateNamespaceStatus(externalContext(ctx, o.cfg), req)
}

func (o *outerNamespaceServer) DeleteNamespace(ctx context.Context, req *goddessv1.DeleteNamespaceRequest) (*goddessv1.DeleteNamespaceReply, error) {
	if o.httpClient != nil {
		return o.httpClient.DeleteNamespace(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.DeleteNamespace(externalContext(ctx, o.cfg), req)
}

func (o *outerNamespaceServer) GetNamespace(ctx context.Context, req *goddessv1.GetNamespaceRequest) (*goddessv1.NamespaceItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetNamespace(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.GetNamespace(externalContext(ctx, o.cfg), req)
}

func (o *outerNamespaceServer) ListNamespace(ctx context.Context, req *goddessv1.ListNamespaceRequest) (*goddessv1.ListNamespaceReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListNamespace(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.ListNamespace(externalContext(ctx, o.cfg), req)
}

func (o *outerNamespaceServer) SelectNamespace(ctx context.Context, req *goddessv1.SelectNamespaceRequest) (*goddessv1.SelectNamespaceReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SelectNamespace(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.SelectNamespace(externalContext(ctx, o.cfg), req)
}

func (o *outerNamespaceServer) GetNamespaceSimple(ctx context.Context, req *goddessv1.GetNamespaceSimpleRequest) (*goddessv1.NamespaceItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetNamespaceSimple(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.GetNamespaceSimple(externalContext(ctx, o.cfg), req)
}
