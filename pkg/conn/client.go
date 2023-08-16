package conn

import (
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	gGrpc "google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
)

type INodeServer interface {
	GetServername() string
	GetNetwork() string
	GetTimeout() *durationpb.Duration
}

type Network = string

const (
	NetworkGrpc  Network = "grpc"
	NetworkHttp  Network = "http"
	NetworkHttps Network = "https"
)

var NetworkError = errors.New("invalid network")

func GetNodeHttpClient(ctx context.Context, s INodeServer, d registry.Discovery) (*http.Client, error) {
	if s.GetNetwork() != NetworkHttp && s.GetNetwork() != NetworkHttps {
		return nil, NetworkError
	}

	hConn, err := http.NewClient(
		ctx,
		http.WithEndpoint("discovery:///"+s.GetServername()),
		http.WithDiscovery(d),
		http.WithTimeout(s.GetTimeout().AsDuration()),
	)

	return hConn, err
}

func GetNodeGrpcClient(ctx context.Context, s INodeServer, d registry.Discovery) (*gGrpc.ClientConn, error) {
	if s.GetNetwork() != NetworkGrpc {
		return nil, NetworkError
	}
	gConn, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint("discovery:///"+s.GetServername()),
		grpc.WithDiscovery(d),
		grpc.WithTimeout(s.GetTimeout().AsDuration()),
	)

	return gConn, err
}
