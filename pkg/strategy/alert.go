package strategy

import (
	"context"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

var _ Alerter = (*Alerting)(nil)

type Alerter interface {
	Eval(ctx context.Context, rule *Rule) ([]*Result, error)
}

type Alerting struct {
	exitCh         chan struct{}
	log            *log.Helper
	datasourceName DatasourceName
	group          *Group
}

func (a *Alerting) Eval(ctx context.Context, rule *Rule) ([]*Result, error) {
	datasource := NewDatasource(a.datasourceName, rule.Endpoint())
	queryResponse, err := datasource.Query(ctx, rule.Expr, time.Now().Unix())
	if err != nil {
		return nil, err
	}

	return queryResponse.Data.Result, nil
}

// NewAlerting 初始化策略告警实例
func NewAlerting(group *Group, datasourceName DatasourceName, logger log.Logger) *Alerting {
	if logger == nil {
		logger = log.DefaultLogger
	}
	a := &Alerting{
		exitCh:         make(chan struct{}),
		datasourceName: datasourceName,
		group:          group,
		log:            log.NewHelper(log.With(logger, "module", "strategy.alerting")),
	}

	return a
}

// BuildDuration 字符串转为api时间
func BuildDuration(duration string) time.Duration {
	durationLen := len(duration)
	if duration == "" || durationLen < 2 {
		return 0
	}
	value, _ := strconv.Atoi(duration[:durationLen-1])
	// 获取字符串最后一个字符
	unit := string(duration[durationLen-1])
	switch unit {
	case "s":
		return time.Duration(value) * time.Second
	case "m":
		return time.Duration(value) * time.Minute
	case "h":
		return time.Duration(value) * time.Hour
	case "d":
		return time.Duration(value) * time.Hour * 24
	default:
		return time.Duration(value) * time.Second
	}
}
