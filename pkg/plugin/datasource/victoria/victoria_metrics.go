package victoria

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/moon-monitor/moon/pkg/plugin/datasource"
)

func New(c datasource.MetricConfig, logger log.Logger) *Victoria {
	return &Victoria{
		c:      c,
		helper: log.NewHelper(log.With(logger, "module", "plugin.datasource.victoria")),
	}
}

type Victoria struct {
	c      datasource.MetricConfig
	helper *log.Helper
}
