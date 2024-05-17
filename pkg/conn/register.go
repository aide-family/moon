package conn

import (
	"github.com/aide-cloud/moon/api"
	"github.com/aide-cloud/moon/api/merr"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// NewRegister 创建一个服务注册实例
func NewRegister(c *api.Registry) (registry.Registrar, error) {
	if types.IsNil(c) {
		return nil, merr.ErrorNotification("registry config is nil")
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
		return nil, merr.ErrorNotification("registry type is not support")
	}
}
