package strategy

import (
	"context"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
)

var _ Alerter = (*Alerting)(nil)

type Alerter interface {
	Eval(ctx context.Context) ([]*Alarm, error)
}

type Alerting struct {
	exitCh         chan struct{}
	log            *log.Helper
	datasourceName DatasourceName
	group          *Group
}

func (a *Alerting) Eval(ctx context.Context) ([]*Alarm, error) {
	eg := new(errgroup.Group)
	eg.SetLimit(100)
	alarmList := NewAlarmList()
	for _, strategyItem := range a.group.Rules {
		strategyInfo := &*strategyItem
		eg.Go(func() error {
			datasource := NewDatasource(a.datasourceName, strategyInfo.Endpoint())
			queryResponse, err := datasource.Query(ctx, strategyInfo.Expr, time.Now().Unix())
			if err != nil {
				return err
			}
			alarmInfo := NewAlarm(a.group, strategyInfo, queryResponse.Data.Result)
			alarmList.Append(alarmInfo)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return alarmList.List(), nil
}

// NewAlerting 初始化策略告警实例
func NewAlerting(group *Group, datasourceName DatasourceName, logger *log.Helper) *Alerting {
	a := &Alerting{
		exitCh:         make(chan struct{}),
		datasourceName: datasourceName,
		group:          group,
		log:            logger,
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
