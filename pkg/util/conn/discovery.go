package conn

import (
	"strings"

	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/types"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type (
	// IEtcdConfig etcd配置
	IEtcdConfig interface {
		GetEndpoints() []string
		GetPassword() string
		GetTimeout() *durationpb.Duration
		GetUsername() string
	}

	// IDiscoveryConfig 服务发现配置
	IDiscoveryConfig interface {
		GetDriver() string
	}

	// discoveryConfig 服务发现配置
	discoveryConfig struct {
		Type string
		Etcd IEtcdConfig
	}

	// DiscoveryConfigBuilderOptions 服务发现配置构建选项
	DiscoveryConfigBuilderOptions func(cfg *discoveryConfig)
)

// NewDiscovery 创建一个服务发现实例
func NewDiscovery(c IDiscoveryConfig, opts ...DiscoveryConfigBuilderOptions) (registry.Discovery, error) {
	if types.IsNil(c) {
		return nil, merr.ErrorNotification("discovery config is nil")
	}
	cfg := &discoveryConfig{
		Type: c.GetDriver(),
	}
	for _, opt := range opts {
		opt(cfg)
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Type)) {
	case "etcd":
		etcdConf := cfg.Etcd
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
		return nil, merr.ErrorNotification("discovery config is not supported")
	}
}

// NewRegister 创建一个服务注册实例
func NewRegister(c IDiscoveryConfig, opts ...DiscoveryConfigBuilderOptions) (registry.Registrar, error) {
	if types.IsNil(c) {
		return nil, merr.ErrorNotification("registry config is nil")
	}
	cfg := &discoveryConfig{
		Type: c.GetDriver(),
	}
	for _, opt := range opts {
		opt(cfg)
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Type)) {
	case "etcd":
		etcdConf := cfg.Etcd
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

// WithDiscoveryConfigEtcd 设置 etcd 配置
func WithDiscoveryConfigEtcd(etcdConf IEtcdConfig) DiscoveryConfigBuilderOptions {
	return func(cfg *discoveryConfig) {
		cfg.Etcd = etcdConf
	}
}
