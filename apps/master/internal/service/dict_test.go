package service

import (
	"sync"
	"testing"

	"golang.org/x/sync/errgroup"
)

func TestSliceAppend(t *testing.T) {
	var eg errgroup.Group
	var list []int
	var lock sync.Mutex
	for i := 0; i < 100; i++ {
		eg.Go(func() error {
			lock.Lock()
			defer lock.Unlock()
			list = append(list, i)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}

	t.Log(len(list))
}
