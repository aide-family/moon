package cache

import (
	"context"

	"k8s.io/client-go/tools/cache"
)

type entry struct {
	indexer cache.Indexer
	reader  Reader
	writer  Writer
}

func (e *entry) Get(ctx context.Context, key string, out any) error {
	return e.reader.Get(ctx, key, out)
}

func (e *entry) List(ctx context.Context, out any) error {
	return e.reader.List(ctx, out)
}

func (e *entry) AddIndexers(indexers cache.Indexers) error {
	return e.indexer.AddIndexers(indexers)
}

func (e *entry) GetIndexer() cache.Indexer {
	return e.indexer
}

func (e *entry) Add(ctx context.Context, objects any) error {
	return e.writer.Add(ctx, objects)
}

func (e *entry) Replace(ctx context.Context, objects any) error {
	return e.writer.Replace(ctx, objects)
}

func (e *entry) Update(ctx context.Context, object any) error {
	return e.writer.Update(ctx, object)
}

func (e *entry) Delete(ctx context.Context, object any) error {
	return e.writer.Delete(ctx, object)
}
