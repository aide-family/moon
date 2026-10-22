package healthcheck

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestChecksNilIsReady(t *testing.T) {
	var checks *Checks
	if err := checks.Check(context.Background()); err != nil {
		t.Fatalf("nil checks should be ready, got %v", err)
	}
}

func TestChecksAllPass(t *testing.T) {
	checks := NewChecks(time.Second)
	checks.Add("postgres", func(context.Context) error { return nil })
	checks.Add("redis", func(context.Context) error { return nil })
	if err := checks.Check(context.Background()); err != nil {
		t.Fatalf("Check() = %v, want nil", err)
	}
}

func TestChecksReportsFailingDependency(t *testing.T) {
	checks := NewChecks(time.Second)
	checks.Add("postgres", func(context.Context) error { return nil })
	checks.Add("redis", func(context.Context) error { return errors.New("down") })
	err := checks.Check(context.Background())
	if err == nil {
		t.Fatal("Check() = nil, want error")
	}
	// 错误里必须带依赖名，否则 /readyz 没法告诉运维是哪一个依赖挂了。
	if got := err.Error(); got != "redis: down" {
		t.Fatalf("Check() error = %q, want %q", got, "redis: down")
	}
}

func TestChecksTimeout(t *testing.T) {
	checks := NewChecks(20 * time.Millisecond)
	checks.Add("slow", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			return nil
		}
	})
	start := time.Now()
	if err := checks.Check(context.Background()); err == nil {
		t.Fatal("Check() = nil, want timeout error")
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("Check() took %v, want it bounded by the check timeout", elapsed)
	}
}

func TestChecksIgnoresNilFunc(t *testing.T) {
	checks := NewChecks(time.Second)
	checks.Add("noop", nil)
	if err := checks.Check(context.Background()); err != nil {
		t.Fatalf("Check() = %v, want nil", err)
	}
}
