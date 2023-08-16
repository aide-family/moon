package conn

import (
	"sync"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	clientV3 "go.etcd.io/etcd/client/v3"
)

type EtcdConfig interface {
	GetEndpoints() []string
}

var (
	dis  registry.Discovery
	reg  *etcd.Registry
	once sync.Once
)

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
	registry := etcd.New(etcdClient)
	once.Do(func() {
		dis = registry
		reg = registry
	})
	return registry
}

func GetDiscovery() registry.Discovery {
	if dis == nil {
		panic("discovery is nil")
	}
	return dis
}

func GetRegistrar() *etcd.Registry {
	if reg == nil {
		panic("registrar is nil")
	}
	return reg
}
