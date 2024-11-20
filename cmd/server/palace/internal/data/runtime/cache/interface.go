package cache

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtime"

	"k8s.io/client-go/tools/cache"
)

// Reader 缓存读取器
type Reader interface {
	runtime.Reader
}

// Writer 缓存写入器
type Writer interface {
	Add(ctx context.Context, object any) error
	Replace(ctx context.Context, objects any) error
	Update(ctx context.Context, object any) error
	Delete(ctx context.Context, object any) error
}

// Informers knows how to create or fetch informers for different
// kinds, and add indices to those informers.  It's safe to call
// GetInformer from multiple threads.
type Informers interface {
	// GetInformer fetches or constructs an informer for the given object that corresponds to a single
	// API kind and resource.
	GetInformer(ctx context.Context, obj any) (Informer, error)

	// GetInformerForKind is similar to GetInformer, except that it takes akind, instead
	// of the underlying object.
	GetInformerForKind(ctx context.Context, kind string) (Informer, error)
}

// Informer 缓存通知器
type Informer interface {
	// AddIndexers adds more indexers to this store.  If you call this after you already have data
	// in the store, the results are undefined.
	AddIndexers(indexers cache.Indexers) error
	// GetIndexer returns the indexer used to watch and store this object.
	GetIndexer() cache.Indexer
}

// Interface 缓存接口
type Interface interface {
	Reader
	Writer
	Informers
}
