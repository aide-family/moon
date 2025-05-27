package server

import (
	"context"
	"sync"
	"time"

	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

var _ transport.Server = (*Ticker)(nil)

func NewTicker(interval time.Duration, task *TickTask, opts ...TickerOption) *Ticker {
	t := &Ticker{
		interval: interval,
		stop:     make(chan struct{}),
		task:     task,
	}
	for _, opt := range opts {
		opt(t)
	}
	if t.helper == nil {
		WithTickerLogger(log.DefaultLogger)(t)
	}
	return t
}

type TickTask struct {
	Fn        func(ctx context.Context, isStop bool) error
	Name      string
	Timeout   time.Duration
	Interval  time.Duration
	Immediate bool
}

type Ticker struct {
	interval  time.Duration
	ticker    *time.Ticker
	stop      chan struct{}
	task      *TickTask
	immediate bool

	helper *log.Helper
}

type TickerOption func(*Ticker)

func WithTickerLogger(logger log.Logger) TickerOption {
	return func(t *Ticker) {
		t.helper = log.NewHelper(log.With(logger, "module", "server.tick"))
	}
}

func WithTickerImmediate(immediate bool) TickerOption {
	return func(t *Ticker) {
		t.immediate = immediate
	}
}

func (t *Ticker) Start(ctx context.Context) error {
	t.ticker = time.NewTicker(t.interval)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				t.helper.Errorw("method", "Start", "panic", err)
			}
		}()
		if t.immediate {
			t.call(ctx, false)
		}
		for {
			select {
			case <-t.ticker.C:
				t.call(ctx, false)
			case <-t.stop:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (t *Ticker) call(ctx context.Context, isStop bool) {
	timeout := t.task.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := t.task.Fn(ctx, isStop); err != nil {
		t.helper.Errorf("execute task %s error: %v", t.task.Name, err)
	}
}

func (t *Ticker) Stop(ctx context.Context) error {
	close(t.stop)
	t.ticker.Stop()
	t.call(ctx, true)
	return nil
}

var _ transport.Server = (*Tickers)(nil)

func NewTickers(opts ...TickersOption) *Tickers {
	t := &Tickers{
		autoID:  uint64(1),
		tickers: make(map[uint64]*Ticker),
		recycle: make([]uint64, 0, 100),
		logger:  log.DefaultLogger,
	}
	for _, opt := range opts {
		opt(t)
	}

	return t
}

type Tickers struct {
	mu      sync.RWMutex
	autoID  uint64
	recycle []uint64
	tickers map[uint64]*Ticker
	logger  log.Logger
}

type TickersOption func(*Tickers)

func WithTickersLogger(logger log.Logger) TickersOption {
	return func(t *Tickers) {
		t.logger = logger
	}
}

func WithTickersTasks(tasks ...*TickTask) TickersOption {
	return func(t *Tickers) {
		for _, task := range tasks {
			t.Add(task.Interval, task)
		}
	}
}

func (t *Tickers) Add(interval time.Duration, task *TickTask) uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	id := t.autoID
	if len(t.recycle) > 0 {
		id = t.recycle[0]
		t.recycle = t.recycle[1:]
	} else {
		t.autoID++
	}
	ticker := NewTicker(interval, task, WithTickerLogger(t.logger), WithTickerImmediate(task.Immediate))
	defer func(ticker *Ticker, ctx context.Context) {
		err := ticker.Start(ctx)
		if err != nil {
			_ = t.logger.Log(log.LevelError, "err", err, "msg", "start ticker err")
		}
	}(ticker, context.Background())
	t.tickers[id] = ticker
	return id
}

func (t *Tickers) Remove(id uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	ticker, ok := t.tickers[id]
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := ticker.Stop(ctx); err != nil {
		_ = t.logger.Log(log.LevelWarn, "err", err, "msg", "stop ticker err")
	}
	delete(t.tickers, id)
	t.recycle = append(t.recycle, id)
}

func (t *Tickers) Start(ctx context.Context) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, ticker := range t.tickers {
		safety.Go(func() error {
			return ticker.Start(ctx)
		})
	}
	return nil
}

func (t *Tickers) Stop(ctx context.Context) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, ticker := range t.tickers {
		safety.Go(func() error {
			return ticker.Stop(ctx)
		})
	}
	t.tickers = make(map[uint64]*Ticker)
	t.recycle = make([]uint64, 0, 100)
	t.autoID = 1
	return nil
}
