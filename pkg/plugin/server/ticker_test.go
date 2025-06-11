package server_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/moon-monitor/moon/pkg/plugin/server"
	"github.com/moon-monitor/moon/pkg/util/timex"
)

// TestNewTicker verifies that TestNewTicker correctly initializes a Ticker with the given interval and task.
func TestNewTicker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	interval := 1 * time.Second
	start := timex.Now()
	task := &server.TickTask{
		Fn: func(ctx context.Context, isStop bool) error {
			if isStop {
				t.Logf("Task stopped")
				return nil
			}
			diff := timex.Now().Sub(start)
			diff = diff.Round(time.Second)
			if diff < interval {
				t.Errorf("Expected task to be executed after %v, but it was executed after %v", interval, diff)
				return fmt.Errorf("task executed after %v", diff)
			}
			t.Logf("Task executed after %v", diff)
			return nil
		},
		Name:    "定时器",
		Timeout: 0,
	}

	ticker := server.NewTicker(interval, task)
	err := ticker.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start timer: %v", err)
	}

	<-ctx.Done()
	ticker.Stop(ctx)
}

func TestTestNewTickers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	list := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		3 * time.Second,
		4 * time.Second,
		5 * time.Second,
	}
	start := timex.Now()
	task := make([]*server.TickTask, len(list))
	for i, v := range list {
		task[i] = &server.TickTask{
			Fn: func(ctx context.Context, isStop bool) error {
				if isStop {
					t.Logf("Task stopped")
					return nil
				}
				diff := timex.Now().Sub(start)
				diff = diff.Round(time.Second)
				if diff < v {
					t.Errorf("Expected task to be executed after %v, but it was executed after %v", v, diff)
					return fmt.Errorf("task executed after %v: %v", v, diff)
				}
				t.Logf("Task executed after %v: %v", v, diff)
				return nil
			},
			Name:     fmt.Sprintf("%v", v),
			Timeout:  v,
			Interval: v,
		}
	}

	tickers := server.NewTickers(server.WithTickersTasks(task...))
	err := tickers.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start timer: %v", err)
	}

	tickers.Add(1*time.Second, &server.TickTask{
		Fn: func(ctx context.Context, isStop bool) error {
			t.Logf("Add 1s Task executed after 1s")
			return nil
		},
		Name:    "1s",
		Timeout: 0,
	})

	tickers.Add(2*time.Second, &server.TickTask{
		Fn: func(ctx context.Context, isStop bool) error {
			t.Logf("Add 2s Task executed after 2s")
			return nil
		},
		Name:    "1s",
		Timeout: 0,
	})

	<-ctx.Done()
	tickers.Stop(ctx)
}
