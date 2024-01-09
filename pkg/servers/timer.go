package servers

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

var _ transport.Server = (*Timer)(nil)

type TimerCall func(ctx context.Context)

type Timer struct {
	call   TimerCall
	ticker *time.Ticker
	logger *log.Helper
}

func (l *Timer) Start(ctx context.Context) (err error) {
	l.logger.WithContext(ctx).Info("[Timer] server starting")
	// 根据ticker的时间间隔，定时执行call
	for {
		select {
		case <-ctx.Done():
			l.close()
			return
		case <-l.ticker.C:
			l.call(ctx)
			return
		}
	}
}

func (l *Timer) close() {
	l.ticker.Stop()
	l.logger.Info("[Timer] server stopped")
}

func (l *Timer) Stop(ctx context.Context) error {
	l.logger.WithContext(ctx).Info("[Timer] server stopping")
	return nil
}

// NewTimer 创建一个定时器
func NewTimer(interval time.Duration, call TimerCall, logger log.Logger) *Timer {
	var ticker = time.NewTicker(interval)
	return &Timer{
		call:   call,
		ticker: ticker,
		logger: log.NewHelper(log.With(logger, "module", "server.timer")),
	}
}
