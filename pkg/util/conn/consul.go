package conn

import (
	palaceconfig "github.com/aide-family/moon/pkg/config"
	"github.com/hashicorp/consul/api"
)

func NewConsul(consulConf *palaceconfig.Consul) (*api.Client, error) {
	return api.NewClient(&api.Config{
		Address: consulConf.GetAddress(),
	})
}
