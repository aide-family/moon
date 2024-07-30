package cache

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/aide-family/moon/pkg/runtime"
	"github.com/aide-family/moon/pkg/runtime/client"
	"k8s.io/client-go/tools/cache"
)

func New(opts Options) (Cache, error) {
	im := newInformersMap(opts.Scheme, *opts.Resync)
	return im, nil
}

var _ Cache = &informerCache{}

type informerCache struct {
	standard *specificInformersMap
	Scheme   *runtime.Scheme
}

func (m *informerCache) Get(context context.Context, key string, out client.Object) error {
	kind, err := m.Scheme.ObjectKind(out)
	if err != nil {
		return err
	}

	started, cache, err := m.standard.Get(context, kind, out)
	if err != nil {
		return err
	}

	if !started {
		return fmt.Errorf("cache not started, %s", kind)
	}
	return cache.Reader.Get(context, key, out)
}

func (m *informerCache) List(context context.Context, out []client.Object) error {
	elementType := reflect.ValueOf(out).Type().Elem()
	single := reflect.New(elementType).Elem().Interface().(runtime.Object)
	kind, err := m.Scheme.ObjectKind(single)
	if err != nil {
		return err
	}

	started, cache, err := m.standard.Get(context, kind, single)
	if err != nil {
		return err
	}

	if !started {
		return fmt.Errorf("cache not started, %s", kind)
	}
	return cache.Reader.List(context, out)
}

func (m *informerCache) GetInformer(ctx context.Context, obj client.Object) (Informer, error) {
	kind, err := m.Scheme.ObjectKind(obj)
	if err != nil {
		return nil, err
	}

	_, i, err := m.standard.Get(ctx, kind, obj)
	if err != nil {
		return nil, err
	}
	return i.Informer, err
}

func (m *informerCache) GetInformerByKind(ctx context.Context, kind string) (Informer, error) {
	obj, err := m.Scheme.New(kind)
	if err != nil {
		return nil, err
	}

	_, i, err := m.standard.Get(ctx, kind, obj)
	if err != nil {
		return nil, err
	}
	return i.Informer, err
}

func (m *informerCache) WaitForCacheSync(ctx context.Context) bool {
	//TODO implement me
	panic("implement me")
}

func newInformersMap(scheme *runtime.Scheme, resync time.Duration) *informerCache {
	return &informerCache{
		standard: newSpecificInformersMap(scheme, resync),
		Scheme:   scheme,
	}
}

func (m *informerCache) Start(ctx context.Context) error {
	go m.standard.Start(ctx)
	<-ctx.Done()
	return nil
}

type MapEntry struct {
	Informer cache.SharedIndexInformer
	Reader   Reader
}

type specificInformersMap struct {
	Scheme *runtime.Scheme

	// informersByKind is the cache of informers keyed by kind
	informersByKind map[string]*MapEntry

	resync time.Duration

	// mu guards access to the map
	mu sync.RWMutex

	stop <-chan struct{}
	// start is true if the informers have been started
	started bool

	startWait chan struct{}
}

func newSpecificInformersMap(scheme *runtime.Scheme, resync time.Duration) *specificInformersMap {
	ip := &specificInformersMap{
		Scheme:          scheme,
		informersByKind: make(map[string]*MapEntry),
		resync:          resync,
		startWait:       make(chan struct{}),
	}
	return ip
}

func (ip *specificInformersMap) Start(ctx context.Context) {
	func() {
		ip.mu.Lock()
		defer ip.mu.Unlock()

		// Set the stop channel so it can be passed to informers that are added later
		ip.stop = ctx.Done()

		for _, informer := range ip.informersByKind {
			go informer.Informer.Run(ctx.Done())
		}

		// Set started to true so we immediately start any informers added later.
		ip.started = true
		close(ip.startWait)
	}()
	<-ctx.Done()
}

func (ip *specificInformersMap) waitForStarted(ctx context.Context) bool {
	select {
	case <-ip.startWait:
		return true
	case <-ctx.Done():
		return false
	}
}

func (ip *specificInformersMap) Get(ctx context.Context, gvk string, obj runtime.Object) (bool, *MapEntry, error) {
	// Return the informer if it is found
	i, started, ok := func() (*MapEntry, bool, bool) {
		ip.mu.RLock()
		defer ip.mu.RUnlock()
		i, ok := ip.informersByKind[gvk]
		return i, ip.started, ok
	}()

	if !ok {
		var err error
		if i, started, err = ip.addInformerToMap(gvk, obj); err != nil {
			return started, nil, err
		}
	}

	if started && !i.Informer.HasSynced() {
		// Wait for it to sync before returning the Informer so that folks don't read from a stale cache.
		if !cache.WaitForCacheSync(ctx.Done(), i.Informer.HasSynced) {
			return started, nil, fmt.Errorf("failed waiting for %T Informer to sync", obj)
		}
	}

	return started, i, nil
}

func (ip *specificInformersMap) addInformerToMap(kind string, obj runtime.Object) (*MapEntry, bool, error) {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	if i, ok := ip.informersByKind[kind]; ok {
		return i, ip.started, nil
	}

	i := &MapEntry{
		Informer: NewSharedIndexInformer(),
		Reader: Reader{indexer: cache.NewIndexer(
			KeyFunc,
			cache.Indexers{
				NameIndex: NameIndexFunc,
			}), kind: kind},
	}
	ip.informersByKind[kind] = i

	if ip.started {
		go i.Informer.Run(ip.stop)
	}
	return i, ip.started, nil
}

func KeyFunc(obj interface{}) (string, error) {
	data := obj.(client.Object)
	return data.GetName(), nil
}

const NameIndex = "name"

func NameIndexFunc(obj interface{}) ([]string, error) {
	data := obj.(client.Object)
	return []string{data.GetName()}, nil
}

var _ Informer = &sharedIndexInformer{}

// TODO: implement it
type sharedIndexInformer struct {
}

func NewSharedIndexInformer() cache.SharedIndexInformer {
	return &sharedIndexInformer{}
}

func (s *sharedIndexInformer) AddEventHandler(handler cache.ResourceEventHandler) {
	//TODO implement me
	panic("implement me")
}

func (s *sharedIndexInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) {
	//TODO implement me
	panic("implement me")
}

func (s *sharedIndexInformer) GetStore() cache.Store {
	//TODO implement me
	panic("implement me")
}

func (s *sharedIndexInformer) GetController() cache.Controller {
	//TODO implement me
	panic("implement me")
}

func (s *sharedIndexInformer) Run(stopCh <-chan struct{}) {
	//TODO implement me
	panic("implement me")
}

func (s *sharedIndexInformer) HasSynced() bool {
	//TODO implement me
	panic("implement me")
}

func (s *sharedIndexInformer) LastSyncResourceVersion() string {
	//TODO implement me
	panic("implement me")
}

func (s *sharedIndexInformer) SetWatchErrorHandler(handler cache.WatchErrorHandler) error {
	//TODO implement me
	panic("implement me")
}

func (s *sharedIndexInformer) AddIndexers(indexers cache.Indexers) error {
	//TODO implement me
	panic("implement me")
}

func (s *sharedIndexInformer) GetIndexer() cache.Indexer {
	//TODO implement me
	panic("implement me")
}
