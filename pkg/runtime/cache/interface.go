package cache

import (
	"context"

	"github.com/aide-family/moon/pkg/runtime/client"
	"k8s.io/client-go/tools/cache"
)

type Cache interface {
	client.Reader
	Informers
}

type Informers interface {
	GetInformer(ctx context.Context, obj client.Object) (Informer, error)
	GetInformerByKind(ctx context.Context, kind string) (Informer, error)
	Start(ctx context.Context) error
	WaitForCacheSync(ctx context.Context) bool
}

type Informer interface {
	AddEventHandler(handler cache.ResourceEventHandler)
	AddIndexers(indexers cache.Indexers) error
	HasSynced() bool
}
