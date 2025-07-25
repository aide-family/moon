package registry

import (
	registryconsul "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	registryetcd "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"

	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/conn"
)

// NewDiscovery Create a service discovery instance
func NewDiscovery(c *config.Registry) (registry.Discovery, error) {
	switch c.GetDriver() {
	case config.RegistryDriver_ETCD:
		etcdConf := c.GetEtcd()
		client, err := conn.NewEtcd(etcdConf)
		if err != nil {
			return nil, err
		}
		return registryetcd.New(client), nil
	case config.RegistryDriver_CONSUL:
		consulConf := c.GetConsul()
		client, err := conn.NewConsul(consulConf)
		if err != nil {
			return nil, err
		}
		return registryconsul.New(client), nil
	default:
		return nil, merr.ErrorInternalServer("discovery config is not supported")
	}
}
