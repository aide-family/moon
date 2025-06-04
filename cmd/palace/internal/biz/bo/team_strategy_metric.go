package bo

import (
	"strings"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/validate"
)

type CreateTeamMetricStrategyParams interface {
	Validate() error
	GetStrategy() do.Strategy
	GetExpr() string
	GetLabels() kv.KeyValues
	GetAnnotations() kv.StringMap
	GetDatasource() []do.DatasourceMetric
	WithStrategy(strategy do.Strategy)
	WithDatasource(datasource []do.DatasourceMetric)
}

type UpdateTeamMetricStrategyParams interface {
	CreateTeamMetricStrategyParams
	GetStrategyMetric() do.StrategyMetric
	WithStrategyMetric(strategyMetric do.StrategyMetric)
}

type SaveTeamMetricStrategyParams struct {
	StrategyID  uint32
	Expr        string
	Labels      kv.KeyValues
	Annotations kv.StringMap
	Datasource  []uint32

	strategyDo       do.Strategy
	datasourceDos    []do.DatasourceMetric
	strategyMetricDo do.StrategyMetric
}

// GetAnnotations implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) GetAnnotations() kv.StringMap {
	return s.Annotations
}

// GetDatasource implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) GetDatasource() []do.DatasourceMetric {
	return s.datasourceDos
}

// GetExpr implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) GetExpr() string {
	return s.Expr
}

// GetLabels implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) GetLabels() kv.KeyValues {
	return s.Labels
}

// GetStrategy implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) GetStrategy() do.Strategy {
	return s.strategyDo
}

// GetStrategyMetric implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) GetStrategyMetric() do.StrategyMetric {
	return s.strategyMetricDo
}

// WithDatasource implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) WithDatasource(datasource []do.DatasourceMetric) {
	s.datasourceDos = datasource
}

// WithStrategy implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) WithStrategy(strategy do.Strategy) {
	s.strategyDo = strategy
}

// WithStrategyMetric implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) WithStrategyMetric(strategyMetric do.StrategyMetric) {
	s.strategyMetricDo = strategyMetric
}

func (s *SaveTeamMetricStrategyParams) Validate() error {
	if s.StrategyID <= 0 {
		return merr.ErrorParams("strategy id is required")
	}
	if validate.IsNil(s.strategyDo) {
		return merr.ErrorParams("strategy is not found")
	}
	if strings.TrimSpace(s.Expr) == "" {
		return merr.ErrorParams("expr is required")
	}
	if len(s.Datasource) == 0 {
		return merr.ErrorParams("datasource is required")
	}
	if validate.IsNil(s.Annotations) {
		return merr.ErrorParams("annotations is required")
	}
	if len(s.Datasource) != len(s.datasourceDos) {
		return merr.ErrorParams("datasource is not found")
	}
	if s.strategyDo.GetStatus().IsEnable() {
		return merr.ErrorBadRequest("enabled strategy cannot modify")
	}
	if validate.IsNotNil(s.strategyMetricDo) {
		if s.strategyMetricDo.GetStrategy().GetID() != s.StrategyID {
			return merr.ErrorParams("strategy metric is not found")
		}
	}

	return nil
}
