package servers

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"prometheus-manager/pkg/runtimehelper"
	"time"
)

type TimerCall func(ctx context.Context) error

type Timer struct {
	call   TimerCall
	ticker *time.Ticker
	stop   chan struct{}
	logger *log.Helper
}

func (l *Timer) Start(ctx context.Context) error {
	l.logger.Info("[Timer] server starting")
	// 根据ticker的时间间隔，定时执行call
	go func() {
		defer runtimehelper.Recover("[Timer] server")
		for {
			select {
			case <-ctx.Done():
				return
			case <-l.ticker.C:
				if err := l.call(ctx); err != nil {
					l.logger.Errorf("[Timer] call error: %v", err)
					return
				}
			case <-l.stop:
				return
			}
		}
	}()

	return nil
}

func (l *Timer) Stop(ctx context.Context) error {
	l.logger.Info("[Timer] server stopping")
	l.ticker.Stop()
	l.stop <- struct{}{}
	return nil
}

// NewTimer 创建一个定时器
func NewTimer(call TimerCall, ticker *time.Ticker, logger log.Logger) *Timer {
	return &Timer{
		call:   call,
		ticker: ticker,
		stop:   make(chan struct{}),
		logger: log.NewHelper(log.With(logger, "module", "server/timer")),
	}
}

var _ transport.Server = (*Timer)(nil)
