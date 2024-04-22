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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	Runnable
	Status
	Healthy
	Info
	Name() string
	Client() client.Client
	Cache() cache.Cache
	ApiExtensions() clientset.Interface
	Kubernetes() kubernetes.Interface
	Dynamic() dynamic.Interface
	RESTMapper() meta.RESTMapper
	Config() *rest.Config
	Discovery() discovery.DiscoveryInterface
}

// Set of cluster clients.
// Used to uniformly manage all cluster clients.
type Set interface {
	Runnable
	Add(clu Client) error
	Remove(name string)
	Client(name string) Client
	Clients() map[string]Client
}

type Status interface {
	RunStatus() Code
	NetStatus() Code
}

// Runnable defines a service interface that can be run and stopped.
// Through this interface, unified and elegant service start and stop can be achieved.
type Runnable interface {
	Start(ctx context.Context) error
	Stop()
}

type ConfigGetter interface {
	Cluster(ctx context.Context, key types.NamespacedName, in *v1beta1.Cluster) error
	Secret(ctx context.Context, key types.NamespacedName, in *v1.Secret) error
}

type Healthy interface {
	Ready(ctx context.Context) (int, error)
	Health(ctx context.Context) (int, error)
}

type Info interface {
	KubernetesVersion() (string, error)
	APIEnablements() ([]v1beta1.APIEnablement, error)
}

type Builder interface {
	Complete() (Client, error)
}

type InitOptions func(Client) error
