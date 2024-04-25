package bo

import (
	"strconv"

	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/agent/strategy"
)

type EvaluateStrategy struct {
	Id          uint32
	Alert       string
	Expr        string
	For         string
	Labels      agent.Labels
	Annotations agent.Annotations

	Datasource agent.Datasource
}

type EvaluateStrategyGroup struct {
	GroupId      uint32
	GroupName    string
	StrategyList []*EvaluateStrategy
}

type EvaluateReqBo struct {
	GroupList []*EvaluateStrategyGroup
}

// ToEvaluateRule 转换为 EvalRule
func (e *EvaluateStrategy) ToEvaluateRule() *strategy.EvalRule {
	if e == nil {
		return nil
	}

	rule := &strategy.EvalRule{
		ID:          strconv.FormatUint(uint64(e.Id), 10),
		Alert:       e.Alert,
		Expr:        e.Expr,
		For:         e.For,
		Labels:      e.Labels,
		Annotations: e.Annotations,
	}

	rule.SetDatasource(e.Datasource)

	return rule
}
