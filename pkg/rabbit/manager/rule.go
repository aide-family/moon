package manager

import (
	"context"
	"github.com/aide-family/moon/pkg/rabbit"
)

var _ rabbit.SuppressorGetter = &SuppressionRuleManager{}

type SuppressionRuleManager struct {
}

func (s *SuppressionRuleManager) Get(context context.Context, id int64) (rabbit.Suppressor, error) {
	//TODO implement me
	panic("implement me")
}
