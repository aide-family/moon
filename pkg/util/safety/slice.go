package safety

import "sync"

type Slice[T any] struct {
	mu sync.RWMutex
	s  []T
}

func NewSlice[T any](size int) *Slice[T] {
	return &Slice[T]{
		s: make([]T, 0, size),
	}
}

func (s *Slice[T]) Append(v ...T) {
	if len(v) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s = append(s.s, v...)
}

func (s *Slice[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.s)
}

func (s *Slice[T]) Get(index int) (v T, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if index < 0 || index >= len(s.s) {
		return
	}
	return s.s[index], true
}

func (s *Slice[T]) Set(index int, v T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < 0 || index >= len(s.s) {
		return false
	}
	s.s[index] = v
	return true
}

func (s *Slice[T]) DeleteFirst() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.s) == 0 {
		return
	}
	s.s = s.s[1:]
}

func (s *Slice[T]) DeleteLast() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.s) == 0 {
		return
	}
	s.s = s.s[:len(s.s)-1]
}

func (s *Slice[T]) Delete(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < 0 || index >= len(s.s) {
		return
	}
	s.s = append(s.s[:index], s.s[index+1:]...)
}

func (s *Slice[T]) Shift() (v T, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.s) == 0 {
		return
	}
	v = s.s[0]
	s.s = s.s[1:]
	return v, true
}

func (s *Slice[T]) Pop() (v T, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.s) == 0 {
		return
	}
	v = s.s[len(s.s)-1]
	s.s = s.s[:len(s.s)-1]
	return v, true
}
