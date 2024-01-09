package strategy

import (
	"context"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/util/slices"
)

var _ Alerter = (*Alerting)(nil)

type Alerter interface {
	Eval(ctx context.Context, group *Group, rule *Rule) ([]*Result, error)
}

type Alerting struct {
	groups []*Group

	CacheAlerter
	exitCh chan struct{}
	log    *log.Helper
}

func (a *Alerting) Eval(ctx context.Context, group *Group, rule *Rule) ([]*Result, error) {
	datasource := NewDatasource(rule.Endpoint())
	queryResponse, err := datasource.Query(ctx, rule.Expr, time.Now().Unix())
	if err != nil {
		return nil, err
	}

	list := slices.To(queryResponse.Data.Result, func(item *Result) *Result {
		// 把规则组属性加入到数据中
		newItem := item
		if newItem.Metric == nil {
			newItem.Metric = make(Metric)
		}
		newItem.Metric.Set(metricGroupName, group.Name)
		newItem.Metric.Set(metricAlert, rule.Alert)
		for k, v := range rule.Labels {
			newItem.Metric.Set(metricRuleLabelPrefix+k, v)
		}
		return newItem
	})

	return list, nil
}

// NewAlerting 初始化策略告警实例
func NewAlerting(logger log.Logger) *Alerting {
	a := &Alerting{
		exitCh: make(chan struct{}),
		log:    log.NewHelper(log.With(logger, "module", "strategy.alerting")),
	}
	if logger == nil {
		a.log = log.NewHelper(log.With(log.DefaultLogger, "module", "strategy.alerting"))
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
