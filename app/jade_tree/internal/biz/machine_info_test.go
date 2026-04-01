package biz

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/jade_tree/internal/biz/bo"
)

type mockMachineInfoRepo struct {
	callCount int32
	delay     time.Duration
}

func (m *mockMachineInfoRepo) Collect(ctx context.Context) (*bo.MachineInfoBo, error) {
	atomic.AddInt32(&m.callCount, 1)
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return &bo.MachineInfoBo{HostName: "cached-host"}, nil
}

func TestMachineInfoCacheHit(t *testing.T) {
	repo := &mockMachineInfoRepo{}
	svc := NewMachineInfo(repo, klog.NewHelper(klog.NewStdLogger(io.Discard)))

	_, err := svc.GetMachineInfo(context.Background())
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	_, err = svc.GetMachineInfo(context.Background())
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if got := atomic.LoadInt32(&repo.callCount); got != 1 {
		t.Fatalf("expected repo called once, got %d", got)
	}
}

func TestMachineInfoConcurrentRefreshSingleFlight(t *testing.T) {
	repo := &mockMachineInfoRepo{delay: 40 * time.Millisecond}
	svc := NewMachineInfo(repo, klog.NewHelper(klog.NewStdLogger(io.Discard)))

	var wg sync.WaitGroup
	const concurrent = 10
	wg.Add(concurrent)
	for i := 0; i < concurrent; i++ {
		go func() {
			defer wg.Done()
			_, err := svc.GetMachineInfo(context.Background())
			if err != nil {
				t.Errorf("get machine info failed: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&repo.callCount); got != 1 {
		t.Fatalf("expected repo called once under concurrency, got %d", got)
	}
}
