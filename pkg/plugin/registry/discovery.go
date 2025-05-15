package registry

import (
	registryetcd "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/moon-monitor/moon/pkg/config"
	"github.com/moon-monitor/moon/pkg/merr"
)

// NewDiscovery Create a service discovery instance
func NewDiscovery(c *config.Registry) (registry.Discovery, error) {
	switch c.GetDriver() {
	case config.RegistryDriver_ETCD:
		etcdConf := c.GetEtcd()
		client, err := clientv3.New(clientv3.Config{
			Endpoints:             etcdConf.GetEndpoints(),
			AutoSyncInterval:      etcdConf.GetAutoSyncInterval().AsDuration(),
			DialTimeout:           etcdConf.GetTimeout().AsDuration(),
			DialKeepAliveTime:     etcdConf.GetDialKeepAliveTime().AsDuration(),
			DialKeepAliveTimeout:  etcdConf.GetDialKeepAliveTimeout().AsDuration(),
			MaxCallSendMsgSize:    int(etcdConf.GetMaxCallSendMsgSize()),
			MaxCallRecvMsgSize:    int(etcdConf.GetMaxCallRecvMsgSize()),
			Username:              etcdConf.GetUsername(),
			Password:              etcdConf.GetPassword(),
			RejectOldCluster:      etcdConf.GetRejectOldCluster(),
			PermitWithoutStream:   etcdConf.GetPermitWithoutStream(),
			MaxUnaryRetries:       uint(etcdConf.GetMaxUnaryRetries()),
			BackoffWaitBetween:    etcdConf.GetBackoffWaitBetween().AsDuration(),
			BackoffJitterFraction: etcdConf.GetBackoffJitterFraction(),
		})
		if err != nil {
			return nil, err
		}
		return registryetcd.New(client), nil
	default:
		return nil, merr.ErrorInternalServerError("discovery config is not supported")
	}
}
