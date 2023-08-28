package conn

import (
	"fmt"
	"sync"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientV3 "go.etcd.io/etcd/client/v3"
)

type IEtcdConfig interface {
	GetEndpoints() []string
	GetKeyFile() string
	GetCertFile() string
	GetTrustedCaFile() string
}

var (
	dis  registry.Discovery
	reg  *etcd.Registry
	once sync.Once
)

// NewETCDClient new etcd client
//
//	@param cfg IEtcdConfig
func NewETCDClient(conf IEtcdConfig) *clientV3.Client {
	etcdConfig, err := genEtcdConfig(conf)
	if err != nil {
		panic(err)
	}
	// new etcd client
	client, err := clientV3.New(*etcdConfig)
	if err != nil {
		panic(err)
	}
	// new reg with etcd client
	return client
}

// genEtcdConfig generates etcd configuration
//
// @param conf IEtcdConfig
func genEtcdConfig(conf IEtcdConfig) (*clientV3.Config, error) {
	config := &clientV3.Config{}
	endpoints := conf.GetEndpoints()
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("etcd endpoints is empty")
	}
	config.Endpoints = endpoints

	ca := conf.GetTrustedCaFile()
	key := conf.GetKeyFile()
	cert := conf.GetCertFile()
	if ca != "" || cert != "" || key != "" {
		tlsInfo := transport.TLSInfo{
			CertFile:      cert,
			KeyFile:       key,
			TrustedCAFile: ca,
		}
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return nil, err
		}
		config.TLS = tlsConfig
	}
	return config, nil
}

// NewETCDRegistrar new etcd registrar
//
//	@param etcdClient *clientV3.Client
func NewETCDRegistrar(etcdClient *clientV3.Client) *etcd.Registry {
	r := etcd.New(etcdClient)
	once.Do(func() {
		dis = r
		reg = r
	})
	return r
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
