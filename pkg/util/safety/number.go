package safety

import "sync"

type Int64 struct {
	mu sync.RWMutex
	n  int64
}

func NewInt64(v int64) *Int64 {
	return &Int64{
		n: v,
	}
}

func (i *Int64) Add(n int64) int64 {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.n += n
	return i.n
}

func (i *Int64) Sub(n int64) int64 {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.n -= n
	return i.n
}

func (i *Int64) Get() int64 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.n
}

func (i *Int64) Set(n int64) int64 {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.n = n
	return i.n
}

func (i *Int64) Inc() int64 {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.n++
	return i.n
}

func (i *Int64) Dec() int64 {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.n--
	return i.n
}

func (i *Int64) Reset() int64 {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.n = 0
	return i.n
}
