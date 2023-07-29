package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	promBizV1 "prometheus-manager/apps/master/internal/biz/prom/v1"
)

type (
	RuleRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *RuleRepo) V1(ctx context.Context) string {
	//TODO implement me
	panic("implement me")
}

var _ promBizV1.IRuleRepo = (*RuleRepo)(nil)

func NewRuleRepo(data *Data, logger log.Logger) *RuleRepo {
	return &RuleRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Rule"))}
}
