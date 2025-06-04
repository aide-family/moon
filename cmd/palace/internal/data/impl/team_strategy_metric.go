package impl

import (
	"context"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/cmd/palace/internal/data/impl/build"
)

func NewTeamStrategyMetricRepo(d *data.Data) repository.TeamStrategyMetric {
	return &teamStrategyMetricRepoImpl{
		Data: d,
	}
}

type teamStrategyMetricRepoImpl struct {
	*data.Data
}

// DeleteByStrategyIds implements repository.TeamStrategyMetric.
func (t *teamStrategyMetricRepoImpl) DeleteByStrategyIds(ctx context.Context, strategyIds ...uint32) error {
	if len(strategyIds) == 0 {
		return nil
	}
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.StrategyMetric
	wrappers := []gen.Condition{
		mutation.TeamID.Eq(teamId),
		mutation.StrategyID.In(strategyIds...),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).Delete()
	return err
}

// Create implements repository.TeamStrategyMetric.
func (t *teamStrategyMetricRepoImpl) Create(ctx context.Context, params bo.CreateTeamMetricStrategyParams) error {
	tx := getTeamBizQuery(ctx, t)

	strategyMetricDo := &team.StrategyMetric{
		StrategyID:  params.GetStrategy().GetID(),
		Expr:        params.GetExpr(),
		Labels:      params.GetLabels(),
		Annotations: params.GetAnnotations(),
	}
	strategyMetricDo.WithContext(ctx)

	if err := tx.StrategyMetric.WithContext(ctx).Create(strategyMetricDo); err != nil {
		return err
	}

	datasourceList := build.ToDatasourceMetrics(ctx, params.GetDatasource())
	if len(datasourceList) > 0 {
		datasource := tx.StrategyMetric.Datasource.WithContext(ctx).Model(strategyMetricDo)
		if err := datasource.Append(datasourceList...); err != nil {
			return err
		}
	}

	return nil
}

// Update implements repository.TeamStrategyMetric.
func (t *teamStrategyMetricRepoImpl) Update(ctx context.Context, params bo.UpdateTeamMetricStrategyParams) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)

	strategyMetricMutation := tx.StrategyMetric
	wrapper := []gen.Condition{
		strategyMetricMutation.StrategyID.Eq(params.GetStrategy().GetID()),
		strategyMetricMutation.TeamID.Eq(teamId),
	}

	strategyMetricMutations := []field.AssignExpr{
		strategyMetricMutation.Expr.Value(params.GetExpr()),
		strategyMetricMutation.Labels.Value(params.GetLabels()),
		strategyMetricMutation.Annotations.Value(params.GetAnnotations()),
	}
	if _, err := strategyMetricMutation.WithContext(ctx).Where(wrapper...).UpdateSimple(strategyMetricMutations...); err != nil {
		return err
	}

	datasourceDos := build.ToDatasourceMetrics(ctx, params.GetDatasource())
	datasourceMutation := tx.StrategyMetric.Datasource.WithContext(ctx).Model(build.ToStrategyMetric(ctx, params.GetStrategyMetric()))
	if len(datasourceDos) > 0 {
		if err := datasourceMutation.Replace(datasourceDos...); err != nil {
			return err
		}
	} else {
		if err := datasourceMutation.Clear(); err != nil {
			return err
		}
	}

	return nil
}

// Delete implements repository.TeamStrategyMetric.
func (t *teamStrategyMetricRepoImpl) Delete(ctx context.Context, strategyMetricId uint32) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)

	strategyMetricMutation := tx.StrategyMetric
	wrapper := []gen.Condition{
		strategyMetricMutation.ID.Eq(strategyMetricId),
		strategyMetricMutation.TeamID.Eq(teamId),
	}

	_, err := strategyMetricMutation.WithContext(ctx).Where(wrapper...).Delete()
	return err
}

// Get implements repository.TeamStrategyMetric.
func (t *teamStrategyMetricRepoImpl) Get(ctx context.Context, strategyMetricId uint32) (do.StrategyMetric, error) {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)

	strategyMetricMutation := tx.StrategyMetric
	wrapper := []gen.Condition{
		strategyMetricMutation.ID.Eq(strategyMetricId),
		strategyMetricMutation.TeamID.Eq(teamId),
	}
	preloads := []field.RelationField{
		strategyMetricMutation.Strategy.RelationField,
		strategyMetricMutation.StrategyMetricRules.AlarmPages,
		strategyMetricMutation.StrategyMetricRules.LabelNotices,
		strategyMetricMutation.StrategyMetricRules.LabelNotices.Notices,
		strategyMetricMutation.Datasource,
		strategyMetricMutation.StrategyMetricRules.Level,
	}
	strategyMetricDo, err := strategyMetricMutation.WithContext(ctx).Where(wrapper...).Preload(preloads...).First()
	if err != nil {
		return nil, strategyMetricNotFound(err)
	}
	return strategyMetricDo, nil
}

func (t *teamStrategyMetricRepoImpl) GetByStrategyId(ctx context.Context, strategyId uint32) (do.StrategyMetric, error) {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)

	strategyMetricMutation := tx.StrategyMetric
	wrapper := []gen.Condition{
		strategyMetricMutation.StrategyID.Eq(strategyId),
		strategyMetricMutation.TeamID.Eq(teamId),
	}
	preloads := []field.RelationField{
		strategyMetricMutation.Strategy.RelationField,
		strategyMetricMutation.StrategyMetricRules.AlarmPages,
		strategyMetricMutation.StrategyMetricRules.LabelNotices,
		strategyMetricMutation.StrategyMetricRules.LabelNotices.Notices,
		strategyMetricMutation.Datasource,
		strategyMetricMutation.StrategyMetricRules.Level,
	}
	strategyMetricDo, err := strategyMetricMutation.WithContext(ctx).Where(wrapper...).Preload(preloads...).First()
	if err != nil {
		return nil, strategyMetricNotFound(err)
	}
	return strategyMetricDo, nil
}
