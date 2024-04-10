package config

import (
	"context"
	"github.com/aide-family/moon/api/cluster/v1beta1"
	clu "github.com/aide-family/moon/app/kubemoon/internal/cluster"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubeConfig struct {
	client.Client
}

func NewConfig(c client.Client) clu.ConfigGetter {
	return &KubeConfig{c}
}

func (c *KubeConfig) Cluster(ctx context.Context, key types.NamespacedName, in *v1beta1.Cluster) error {
	return c.Get(ctx, key, in)
}

func (c *KubeConfig) Secret(ctx context.Context, key types.NamespacedName, in *v1.Secret) error {
	return c.Get(ctx, key, in)
}
