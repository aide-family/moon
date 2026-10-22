package lock

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestAcquireRenewRelease(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	explicit := addr != ""
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		if !explicit {
			t.Skipf("redis %s is not available: %v", addr, err)
		}
		t.Fatalf("redis %s: %v", addr, err)
	}

	key := "alioth:lock:test:" + randomToken()
	t.Cleanup(func() { _ = rdb.Del(context.Background(), key).Err() })

	first := NewClient(rdb, "proc-a")
	second := NewClient(rdb, "proc-b")

	hold, err := first.Acquire(ctx, key, 2*time.Second)
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	if _, err := second.Acquire(ctx, key, 2*time.Second); !errors.Is(err, ErrNotAcquired) {
		t.Fatalf("second acquire err = %v, want ErrNotAcquired", err)
	}

	before, err := rdb.PTTL(ctx, key).Result()
	if err != nil {
		t.Fatalf("pttl: %v", err)
	}
	if err := hold.Renew(ctx, 30*time.Second); err != nil {
		t.Fatalf("renew: %v", err)
	}
	after, err := rdb.PTTL(ctx, key).Result()
	if err != nil {
		t.Fatalf("pttl after renew: %v", err)
	}
	if after <= before {
		t.Fatalf("ttl after renew = %s, before = %s", after, before)
	}

	wrong := &Hold{client: second, key: key, token: "proc-b:not-the-token"}
	if err := wrong.Release(ctx); !errors.Is(err, ErrTokenMismatch) {
		t.Fatalf("wrong token release err = %v, want ErrTokenMismatch", err)
	}
	exists, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		t.Fatalf("exists: %v", err)
	}
	if exists != 1 {
		t.Fatal("wrong token deleted the key")
	}

	if err := hold.Release(ctx); err != nil {
		t.Fatalf("owner release: %v", err)
	}
	exists, err = rdb.Exists(ctx, key).Result()
	if err != nil {
		t.Fatalf("exists after release: %v", err)
	}
	if exists != 0 {
		t.Fatal("owner release left the key")
	}
}

func TestRedisUnavailable(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:1",
		DialTimeout: 200 * time.Millisecond,
		MaxRetries:  -1,
	})
	defer rdb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := NewClient(rdb, "down").Acquire(ctx, "alioth:lock:test:down", time.Second)
	if err == nil {
		t.Fatal("acquire against a closed redis returned success")
	}
	if errors.Is(err, ErrNotAcquired) {
		t.Fatal("redis outage was reported as a lock conflict")
	}
}
