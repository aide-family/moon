package client

import (
	"context"
	"fmt"
	"github.com/aide-family/moon/api/cluster/v1beta1"
	clu "github.com/aide-family/moon/app/kubemoon/internal/cluster"
	restutil "github.com/aide-family/moon/pkg/util/rest"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

type Builder struct {
	clusterName string
	config      clu.ConfigGetter
	scheme      *runtime.Scheme
	options     []InitOptions
}

func By(config clu.ConfigGetter) *Builder {
	return &Builder{config: config}
}

func (b *Builder) WithScheme(scheme *runtime.Scheme) *Builder {
	b.scheme = scheme
	return b
}

func (b *Builder) WithOptions(opts ...InitOptions) *Builder {
	b.options = opts
	return b
}

func (b *Builder) Named(clusterName string) *Builder {
	b.clusterName = clusterName
	return b
}

func (b *Builder) Complete() (clu.Client, error) {
	if b.scheme == nil {
		return nil, fmt.Errorf("must provide a non-nil scheme")
	}
	clusterCR, err := b.loadClusterCR()
	if err != nil {
		return nil, fmt.Errorf("failed to load clientx resource: %s", err)
	}
	config, err := b.loadConfig(clusterCR.Spec.Connect)
	if err != nil {
		return nil, fmt.Errorf("failed to load client rest config: %s", err)
	}
	cluster, err := New(config, b.scheme, b.options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientx: %s", err)
	}
	if clusterCR.Spec.Disabled {
		cluster.Disable()
	}
	return cluster, nil
}

func (b *Builder) loadClusterCR() (*v1beta1.Cluster, error) {
	if b.config == nil {
		return nil, fmt.Errorf("must provide a non-nil config client")
	}
	if len(b.clusterName) == 0 {
		return nil, fmt.Errorf("must provide a non-empty clientx name")
	}
	return b.clusterGetter(b.clusterName)
}

func (b *Builder) loadConfig(connect v1beta1.ConnectConfig) (*rest.Config, error) {
	return restutil.BuildConfig(b.clusterName, connect, b.secretGetter)
}

func (b *Builder) clusterGetter(name string) (*v1beta1.Cluster, error) {
	cluster := &v1beta1.Cluster{}
	err := b.config.Cluster(context.Background(), types.NamespacedName{Name: name}, cluster)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (b *Builder) secretGetter(key types.NamespacedName) (*v1.Secret, error) {
	secret := &v1.Secret{}
	err := b.config.Secret(context.TODO(), key, secret)
	return secret, err
}
