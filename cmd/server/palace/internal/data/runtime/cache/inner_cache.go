package cache

import (
	"context"
	"fmt"
	"sync"

	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtime"

	"k8s.io/client-go/tools/cache"
)

// innerCache 内部缓存
type innerCache struct {
	cache  *innerCacheMap
	scheme *runtime.Scheme
}

// NewCache 创建缓存
func NewCache(scheme *runtime.Scheme) Interface {
	return &innerCache{
		cache:  newInnerCacheMap(scheme),
		scheme: scheme,
	}
}

// GetInformer 获取信息器
func (i *innerCache) GetInformer(ctx context.Context, obj any) (Informer, error) {
	kind, err := i.scheme.ObjectKind(obj)
	if err != nil {
		return nil, err
	}
	return i.cache.Get(ctx, kind, obj)
}

// GetInformerForKind 获取信息器
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

// Get 获取数据
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

// List 获取数据列表
func (i *innerCache) List(ctx context.Context, out any) error {
	e, err := i.getEntry(ctx, out)
	if err != nil {
		return err
	}
	return e.List(ctx, out)
}

// Add 添加数据
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

// Replace 替换数据
func (i *innerCache) Replace(ctx context.Context, objects any) error {
	e, err := i.getEntry(ctx, objects)
	if err != nil {
		return err
	}
	return e.Replace(ctx, objects)
}

// Update 更新数据
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

// Delete 删除数据
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

// getEntry 获取条目
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

// innerCacheMap 内部缓存映射
type innerCacheMap struct {
	kindCacheMap map[string]*entry
	mu           sync.RWMutex
	scheme       *runtime.Scheme
}

// newInnerCacheMap 创建内部缓存映射
func newInnerCacheMap(scheme *runtime.Scheme) *innerCacheMap {
	return &innerCacheMap{
		kindCacheMap: make(map[string]*entry),
		scheme:       scheme,
	}
}

// Get 获取条目
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

// addInformerToMap 添加信息器到映射
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
