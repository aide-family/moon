package watch_test

import (
	"context"
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/watch"
)

func TestNewWatcher(t *testing.T) {
	w := watch.NewWatcher()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	w.Start(ctx)

	time.Sleep(10 * time.Second)
	w.Stop(context.Background())
}
