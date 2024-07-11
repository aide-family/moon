package repoimpl

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/go-kratos/kratos/v2/log"
)

func NewStrategyRepository(data *data.Data) repository.Strategy {
	return &strategyRepositoryImpl{data: data}
}

type strategyRepositoryImpl struct {
	data *data.Data
}

func (s *strategyRepositoryImpl) Save(_ context.Context, strategies []*bo.Strategy) error {
	queue := s.data.GetStrategyQueue()
	go func() {
		defer after.RecoverX()
		for _, strategyItem := range strategies {
			if err := queue.Push(strategyItem.Message()); err != nil {
				log.Errorw("method", "queue.push", "error", err)
			}
		}
	}()

	return nil
}

func (s *strategyRepositoryImpl) Eval(ctx context.Context, strategy *bo.Strategy) (*bo.Alarm, error) {
	datasourceList := strategy.Datasource
	if len(datasourceList) == 0 {
		return nil, merr.ErrorNotification("datasource is empty")
	}
	category := datasourceList[0].Category
	var alarmInfo bo.Alarm
	datasourceCliList := make([]datasource.Datasource, 0, len(datasourceList))
	for _, datasourceItem := range datasourceList {
		if datasourceItem.Category != category {
			log.Warnw("method", "Eval", "error", "datasource category is not same")
			continue
		}
		// TODO append datasource client
	}
	var points []*datasource.Point
	for _, cli := range datasourceCliList {
		// TODO async eval
		evalPoints, err := cli.Eval(ctx, strategy.Expr)
		if err != nil {
			log.Warnw("method", "Eval", "error", err)
		}
		points = append(points, evalPoints...)
	}
	// TODO points to alarm info
	return &alarmInfo, nil
}
