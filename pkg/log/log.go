package log

import (
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"

	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/hello"
)

func New(isDev bool, cfg *config.Log) (logger log.Logger) {
	switch cfg.GetDriver() {
	case config.Log_SUGARED:
		logger = newSugaredLogger(isDev, cfg.GetLevel(), cfg.GetSugared())
	default:
		logger = log.NewStdLogger(os.Stdout)
	}
	env := hello.GetEnv()
	return log.With(logger,
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", env.ID(),
		"service.name", env.Name(),
		"service.version", env.Version(),
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
}
