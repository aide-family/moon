package fake

import (
	"context"
	"github.com/aide-family/moon/api/cluster/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	client.Client
}

func NewConfig(c client.Client) *Config {
	return &Config{c}
}

func (c *Config) Cluster(ctx context.Context, key types.NamespacedName, in *v1beta1.Cluster) error {
	return c.Get(ctx, key, in)
}

func (c *Config) Secret(ctx context.Context, key types.NamespacedName, in *v1.Secret) error {
	return c.Get(ctx, key, in)
}
