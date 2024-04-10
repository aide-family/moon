package cluster

import (
	"context"
	"github.com/aide-family/moon/api/cluster/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	Runnable
	Status
	Name() string
	Client() client.Client
	Cache() cache.Cache
	ApiExtensions() clientset.Interface
	Dynamic() dynamic.Interface
	RESTMapper() meta.RESTMapper
	Config() *rest.Config
	Discovery() discovery.DiscoveryInterface
}

type Set interface {
	Runnable
	Add(clu Client) error
	Remove(name string)
	Cluster(name string) Client
	Clusters() map[string]Client
}

type Status interface {
	Status() Code
	Enable()
	Disable()
}

type Runnable interface {
	Start(ctx context.Context) error
	Stop()
}

type ConfigGetter interface {
	Cluster(ctx context.Context, key types.NamespacedName, in *v1beta1.Cluster) error
	Secret(ctx context.Context, key types.NamespacedName, in *v1.Secret) error
}
