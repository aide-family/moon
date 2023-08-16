package conn

import (
	"context"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	clientV3 "go.etcd.io/etcd/client/v3"
	gGrpc "google.golang.org/grpc"
	"time"
)

type EtcdConfig interface {
	GetEndpoints() []string
}

// NewETCDClient new etcd client
//
//	@param cfg EtcdConfig
func NewETCDClient(conf EtcdConfig) *clientV3.Client {
	endpoints := conf.GetEndpoints()
	if len(endpoints) == 0 {
		panic("etcd endpoints is empty")
	}
	// new etcd client
	client, err := clientV3.New(clientV3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		panic(err)
	}
	// new reg with etcd client
	return client
}

// NewETCDRegistrar new etcd registrar
//
//	@param etcdClient *clientV3.Client
func NewETCDRegistrar(etcdClient *clientV3.Client) *etcd.Registry {
	return etcd.New(etcdClient)
}

func GetGRPCConn(client *clientV3.Client, serverName string) (*gGrpc.ClientConn, error) {
	ctx, cel := context.WithTimeout(context.Background(), time.Second*60)
	dis := etcd.New(client, etcd.Context(ctx))
	endpoint := "discovery:///" + serverName

	defer cel()
	conn, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(endpoint),
		grpc.WithDiscovery(dis),
		grpc.WithTimeout(60*time.Second),
	)
	return conn, err
}
