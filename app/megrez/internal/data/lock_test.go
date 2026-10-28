package data

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/aide-family/moon/app/megrez/internal/conf"
	"github.com/aide-family/moon/pkg/lock"
)

func TestTwoClaimantsOneEvalLock(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip(err)
	}
	first, _, err := NewLocker(&conf.Redis{Addr: addr}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := NewLocker(&conf.Redis{Addr: addr}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	key := "alioth:lock:eval:test-" + time.Now().Format("150405.000")
	t.Cleanup(func() { _ = rdb.Del(context.Background(), key).Err() })
	hold, err := first.Acquire(ctx, key, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer hold.Release(context.Background())
	if _, err := second.Acquire(ctx, key, 30*time.Second); !errors.Is(err, lock.ErrNotAcquired) {
		t.Fatalf("second acquire = %v", err)
	}
}
