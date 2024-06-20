package conn

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// NewDiscovery 创建一个服务发现实例
func NewDiscovery(c *api.Discovery) (registry.Discovery, error) {
	if types.IsNil(c) {
		return nil, merr.ErrorDependencyErr("discovery config is nil")
	}
	switch c.GetType() {
	case "etcd", "ETCD":
		etcdConf := c.GetEtcd()
		// new etcd client
		client, err := clientv3.New(clientv3.Config{
			Endpoints:   etcdConf.GetEndpoints(),
			DialTimeout: etcdConf.GetTimeout().AsDuration(),
			Username:    etcdConf.GetUsername(),
			Password:    etcdConf.GetPassword(),
		})
		if err != nil {
			return nil, err
		}
		return etcd.New(client), nil
	default:
		return nil, merr.ErrorDependencyErr("discovery config is not supported")
	}
}
