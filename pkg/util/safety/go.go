package safety

import (
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
)

var (
	eg          = new(errgroup.Group)
	egLimit     int
	egLimitOnce sync.Once
)

func SetEgLimit(limit int) {
	egLimitOnce.Do(func() {
		egLimit = limit
		eg.SetLimit(egLimit)
	})
}

func Go(f func() error) {
	eg.Go(func() error {
		defer func() {
			if r := recover(); r != nil {
				log.Errorw("msg", "panic in safety.Go", "error", r)
			}
		}()
		return f()
	})
}

func Wait() error {
	return eg.Wait()
}
