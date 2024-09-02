package cache

import (
	"context"
	"fmt"
	"sync"

	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtime"

	"k8s.io/client-go/tools/cache"
)

type innerCache struct {
	cache  *innerCacheMap
	scheme *runtime.Scheme
}

func NewCache(scheme *runtime.Scheme) Interface {
	return &innerCache{
		cache:  newInnerCacheMap(scheme),
		scheme: scheme,
	}
}

func (i *innerCache) GetInformer(ctx context.Context, obj any) (Informer, error) {
	kind, err := i.scheme.ObjectKind(obj)
	if err != nil {
		return nil, err
	}
	return i.cache.Get(ctx, kind, obj)
}

func (i *innerCache) GetInformerForKind(ctx context.Context, kind string) (Informer, error) {
	if !i.scheme.Recognizes(kind) {
		return nil, fmt.Errorf("kind %s not found", kind)
	}
	obj, err := i.scheme.New(kind)
	if err != nil {
		return nil, err
	}
	return i.cache.Get(ctx, kind, obj)
}

func (i *innerCache) Get(ctx context.Context, key string, out any) error {
	kind, err := i.scheme.ObjectKind(out)
	if err != nil {
		return err
	}
	e, err := i.cache.Get(ctx, kind, out)
	if err != nil {
		return err
	}
	return e.Get(ctx, key, out)
}

func (i *innerCache) List(ctx context.Context, out any) error {
	e, err := i.getEntry(ctx, out)
	if err != nil {
		return err
	}
	return e.List(ctx, out)
}

func (i *innerCache) Add(ctx context.Context, object any) error {
	kind, err := i.scheme.ObjectKind(object)
	if err != nil {
		return err
	}
	e, err := i.cache.Get(ctx, kind, object)
	if err != nil {
		return err
	}
	return e.Add(ctx, object)
}

func (i *innerCache) Replace(ctx context.Context, objects any) error {
	e, err := i.getEntry(ctx, objects)
	if err != nil {
		return err
	}
	return e.Replace(ctx, objects)
}

func (i *innerCache) Update(ctx context.Context, object any) error {
	kind, err := i.scheme.ObjectKind(object)
	if err != nil {
		return err
	}
	e, err := i.cache.Get(ctx, kind, object)
	if err != nil {
		return err
	}
	return e.Update(ctx, object)
}

func (i *innerCache) Delete(ctx context.Context, object any) error {
	kind, err := i.scheme.ObjectKind(object)
	if err != nil {
		return err
	}
	e, err := i.cache.Get(ctx, kind, object)
	if err != nil {
		return err
	}
	return e.Delete(ctx, object)
}

func (i *innerCache) getEntry(ctx context.Context, objects any) (*entry, error) {
	kind, err := i.scheme.ObjectsKind(objects)
	if err != nil {
		return nil, err
	}
	obj, err := i.scheme.New(kind)
	if err != nil {
		return nil, err
	}
	return i.cache.Get(ctx, kind, obj)
}

type innerCacheMap struct {
	kindCacheMap map[string]*entry
	mu           sync.RWMutex
	scheme       *runtime.Scheme
}

func newInnerCacheMap(scheme *runtime.Scheme) *innerCacheMap {
	return &innerCacheMap{
		kindCacheMap: make(map[string]*entry),
		scheme:       scheme,
	}
}

func (ic *innerCacheMap) Get(_ context.Context, kind string, obj any) (*entry, error) {
	// Return the informer if it is found
	e, ok := func() (*entry, bool) {
		ic.mu.RLock()
		defer ic.mu.RUnlock()
		e, ok := ic.kindCacheMap[kind]
		return e, ok
	}()

	if !ok {
		var err error
		if e, err = ic.addInformerToMap(kind, obj); err != nil {
			return nil, err
		}
	}

	return e, nil
}

func (ic *innerCacheMap) addInformerToMap(kind string, obj any) (*entry, error) {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	if e, ok := ic.kindCacheMap[kind]; ok {
		return e, nil
	}
	keyFunc, err := ic.scheme.ObjectKeyFunc(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to get key func: %v", err)
	}

	indexer := cache.NewIndexer(keyFunc, cache.Indexers{})
	e := &entry{
		indexer: indexer,
		reader: &reader{
			indexer: indexer,
			kind:    kind,
		},
		writer: &writer{
			indexer: indexer,
		},
	}
	ic.kindCacheMap[kind] = e

	return e, nil
}
