package biz

import (
	"context"

	"github.com/aide-family/moon/app/prom_agent/internal/biz/bo"
	"github.com/aide-family/moon/pkg/agent"
	"github.com/go-kratos/kratos/v2/log"
)

type EvaluateBiz struct {
	log *log.Helper
}

func NewEvaluateBiz(logger log.Logger) *EvaluateBiz {
	return &EvaluateBiz{log: log.NewHelper(log.With(logger, "module", "biz.evaluate"))}
}

// EvaluateV2 告警事件产生
func (b *EvaluateBiz) EvaluateV2(ctx context.Context, req *bo.EvaluateReqBo) ([]*agent.Alarm, error) {
	alarms := make([]*agent.Alarm, 0, 100)
	for _, group := range req.GroupList {
		for _, strategy := range group.StrategyList {
			alarm, err := strategy.ToEvaluateRule().Eval(ctx)
			if err != nil {
				b.log.Errorf("strategy eval error: %v", err)
				continue
			}
			if alarm == nil {
				continue
			}
			alarms = append(alarms, alarm)
		}
	}
	return alarms, nil
}
