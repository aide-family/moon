package log

import (
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/moon-monitor/moon/pkg/hello"

	"github.com/moon-monitor/moon/pkg/config"
)

func New(isDev bool, cfg *config.Log) (logger log.Logger, err error) {
	defer func() {
		env := hello.GetEnv()
		logger = log.With(log.NewStdLogger(os.Stdout),
			"ts", log.DefaultTimestamp,
			"caller", log.DefaultCaller,
			"service.id", env.ID(),
			"service.name", env.Name(),
			"service.version", env.Version(),
			"trace.id", tracing.TraceID(),
			"span.id", tracing.SpanID(),
		)
	}()
	switch cfg.GetDriver() {
	case config.Log_SUGARED:
		return NewSugaredLogger(isDev, cfg.GetLevel(), cfg.GetSugared())
	default:
		return log.NewStdLogger(os.Stdout), nil
	}
}
