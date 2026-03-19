package namespacev1

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	namespacedomain "github.com/aide-family/goddess/domain/namespace"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func init() {
	namespacedomain.RegisterNamespaceV1Factory(config.DomainConfig_OUTER, NewOuterNamespace)
}

// NewOuterNamespace creates a namespace client that calls a remote goddess (OUTER driver).
func NewOuterNamespace(c *config.DomainConfig) (goddessv1.NamespaceServer, func() error, error) {
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
	initCfg := connect.NewDefaultConfig("goddess.namespace", outer.GetAddress(), timeout, outer.GetProtocol().String())

	switch outer.GetProtocol().String() {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewNamespaceHTTPClient(httpClient)
		return &outerNamespaceServer{httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewNamespaceClient(grpcConn)
		return &outerNamespaceServer{grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", outer.GetProtocol().String())
	}
}

type outerNamespaceServer struct {
	goddessv1.UnimplementedNamespaceServer

	httpClient goddessv1.NamespaceHTTPClient
	grpcClient goddessv1.NamespaceClient
}

func (o *outerNamespaceServer) CreateNamespace(ctx context.Context, req *goddessv1.CreateNamespaceRequest) (*goddessv1.CreateNamespaceReply, error) {
	if o.httpClient != nil {
		return o.httpClient.CreateNamespace(ctx, req)
	}
	return o.grpcClient.CreateNamespace(ctx, req)
}

func (o *outerNamespaceServer) UpdateNamespace(ctx context.Context, req *goddessv1.UpdateNamespaceRequest) (*goddessv1.UpdateNamespaceReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateNamespace(ctx, req)
	}
	return o.grpcClient.UpdateNamespace(ctx, req)
}

func (o *outerNamespaceServer) UpdateNamespaceStatus(ctx context.Context, req *goddessv1.UpdateNamespaceStatusRequest) (*goddessv1.UpdateNamespaceStatusReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateNamespaceStatus(ctx, req)
	}
	return o.grpcClient.UpdateNamespaceStatus(ctx, req)
}

func (o *outerNamespaceServer) DeleteNamespace(ctx context.Context, req *goddessv1.DeleteNamespaceRequest) (*goddessv1.DeleteNamespaceReply, error) {
	if o.httpClient != nil {
		return o.httpClient.DeleteNamespace(ctx, req)
	}
	return o.grpcClient.DeleteNamespace(ctx, req)
}

func (o *outerNamespaceServer) GetNamespace(ctx context.Context, req *goddessv1.GetNamespaceRequest) (*goddessv1.NamespaceItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetNamespace(ctx, req)
	}
	return o.grpcClient.GetNamespace(ctx, req)
}

func (o *outerNamespaceServer) ListNamespace(ctx context.Context, req *goddessv1.ListNamespaceRequest) (*goddessv1.ListNamespaceReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListNamespace(ctx, req)
	}
	return o.grpcClient.ListNamespace(ctx, req)
}

func (o *outerNamespaceServer) SelectNamespace(ctx context.Context, req *goddessv1.SelectNamespaceRequest) (*goddessv1.SelectNamespaceReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SelectNamespace(ctx, req)
	}
	return o.grpcClient.SelectNamespace(ctx, req)
}

func (o *outerNamespaceServer) GetNamespaceSimple(ctx context.Context, req *goddessv1.GetNamespaceSimpleRequest) (*goddessv1.NamespaceItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetNamespaceSimple(ctx, req)
	}
	return o.grpcClient.GetNamespaceSimple(ctx, req)
}
