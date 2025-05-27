package load

import (
	"github.com/go-kratos/kratos/contrib/config/consul/v2"
	etcd "github.com/go-kratos/kratos/contrib/config/etcd/v2"
	"github.com/go-kratos/kratos/v2/config"
	"google.golang.org/protobuf/proto"

	palaceconfig "github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/validate"
)

type ConfigSource interface {
	proto.Message
	GetConfigSource() *palaceconfig.ConfigSource
}

func loadByConfigSource(c ConfigSource) error {
	cs := c.GetConfigSource()
	if validate.IsNil(cs) {
		return nil
	}
	var (
		source config.Source
	)
	switch cs.GetDriver() {
	case palaceconfig.RegistryDriver_CONSUL:
		consulClient, err := conn.NewConsul(cs.GetConsul())
		if err != nil {
			return err
		}
		source, err = consul.New(consulClient, consul.WithPath("app/cart/configs/"))
		if err != nil {
			return err
		}
	case palaceconfig.RegistryDriver_ETCD:
		etcdClient, err := conn.NewEtcd(cs.GetEtcd())
		if err != nil {
			return err
		}
		source, err = etcd.New(etcdClient)
		if err != nil {
			return err
		}
	default:
		return nil
	}

	cfg := config.New(config.WithSource(source))
	if err := cfg.Load(); err != nil {
		return err
	}

	if err := cfg.Scan(c); err != nil {
		return err
	}

	return nil
}
