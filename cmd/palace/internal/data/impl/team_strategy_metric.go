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
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
)

func NewTeamStrategyMetricRepo(d *data.Data) repository.TeamStrategyMetric {
	return &teamStrategyMetricImpl{
		Data: d,
	}
}

type teamStrategyMetricImpl struct {
	*data.Data
}

// DeleteUnUsedLevels implements repository.TeamStrategyMetric.
func (t *teamStrategyMetricImpl) DeleteUnUsedLevels(ctx context.Context, params *bo.DeleteUnUsedLevelsParams) error {
	if params.StrategyMetricID <= 0 || len(params.RuleIds) == 0 {
		return nil
	}
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	strategyMetricRuleMutation := tx.StrategyMetricRule
	wrapper := []gen.Condition{
		strategyMetricRuleMutation.StrategyMetricID.Eq(params.StrategyMetricID),
		strategyMetricRuleMutation.TeamID.Eq(teamId),
		strategyMetricRuleMutation.ID.In(params.RuleIds...),
	}
	_, err := strategyMetricRuleMutation.WithContext(ctx).Where(wrapper...).Delete()
	return err
}

// FindLevels implements repository.TeamStrategyMetric.
func (t *teamStrategyMetricImpl) FindLevels(ctx context.Context, params *bo.FindTeamMetricStrategyLevelsParams) ([]do.StrategyMetricRule, error) {
	if params.StrategyMetricID <= 0 {
		return nil, nil
	}
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	strategyMetricRuleMutation := tx.StrategyMetricRule
	wrapper := []gen.Condition{
		strategyMetricRuleMutation.StrategyMetricID.Eq(params.StrategyMetricID),
		strategyMetricRuleMutation.TeamID.Eq(teamId),
	}
	if len(params.RuleIds) > 0 {
		wrapper = append(wrapper, strategyMetricRuleMutation.ID.In(params.RuleIds...))
	}
	strategyMetricRuleDos, err := strategyMetricRuleMutation.WithContext(ctx).Where(wrapper...).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(strategyMetricRuleDos, func(v *team.StrategyMetricRule) do.StrategyMetricRule {
		return v
	}), nil
}

// Create implements repository.TeamStrategyMetric.
func (t *teamStrategyMetricImpl) Create(ctx context.Context, params bo.CreateTeamMetricStrategyParams) (do.StrategyMetric, error) {
	tx := getTeamBizQuery(ctx, t)

	strategyMetricDo := &team.StrategyMetric{
		StrategyID:  params.GetStrategy().GetID(),
		Strategy:    build.ToStrategy(ctx, params.GetStrategy()),
		Expr:        params.GetExpr(),
		Labels:      params.GetLabels(),
		Annotations: params.GetAnnotations(),
		Datasource:  build.ToDatasourceMetrics(ctx, params.GetDatasource()),
	}
	strategyMetricDo.WithContext(ctx)

	if err := tx.StrategyMetric.WithContext(ctx).Create(strategyMetricDo); err != nil {
		return nil, err
	}

	if len(strategyMetricDo.Datasource) > 0 {
		datasource := tx.StrategyMetric.Datasource.WithContext(ctx).Model(strategyMetricDo)
		if err := datasource.Append(strategyMetricDo.Datasource...); err != nil {
			return nil, err
		}
	}

	return strategyMetricDo, nil
}

// Update implements repository.TeamStrategyMetric.
func (t *teamStrategyMetricImpl) Update(ctx context.Context, params bo.UpdateTeamMetricStrategyParams) (do.StrategyMetric, error) {
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
		return nil, err
	}
	strategyMetricDo, err := t.Get(ctx, &bo.OperateTeamStrategyParams{
		StrategyId: params.GetStrategyMetric().GetID(),
	})
	if err != nil {
		return nil, err
	}

	datasourceDos := build.ToDatasourceMetrics(ctx, params.GetDatasource())
	datasourceMutation := tx.StrategyMetric.Datasource.WithContext(ctx).Model(build.ToStrategyMetric(ctx, strategyMetricDo))
	if len(datasourceDos) > 0 {
		if err := datasourceMutation.Replace(datasourceDos...); err != nil {
			return nil, err
		}
	} else {
		if err := datasourceMutation.Clear(); err != nil {
			return nil, err
		}
	}

	return strategyMetricDo, nil
}

// UpdateLevels implements repository.TeamStrategyMetric.
func (t *teamStrategyMetricImpl) UpdateLevels(ctx context.Context, params bo.SaveTeamMetricStrategyLevels) ([]do.StrategyMetricRule, error) {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)

	strategyMetricRuleMutation := tx.StrategyMetricRule
	wrapper := []gen.Condition{
		strategyMetricRuleMutation.StrategyMetricID.Eq(params.GetStrategyMetric().GetID()),
		strategyMetricRuleMutation.TeamID.Eq(teamId),
	}
	strategyMetricRuleDos := make([]do.StrategyMetricRule, 0, len(params.GetLevels()))
	for _, rule := range params.GetLevels() {
		mutations := []field.AssignExpr{
			strategyMetricRuleMutation.LevelID.Value(rule.GetLevel().GetID()),
			strategyMetricRuleMutation.SampleMode.Value(rule.GetSampleMode().GetValue()),
			strategyMetricRuleMutation.Total.Value(rule.GetTotal()),
			strategyMetricRuleMutation.Condition.Value(rule.GetCondition().GetValue()),
			strategyMetricRuleMutation.Values.Value(team.Values(rule.GetValues())),
			strategyMetricRuleMutation.Status.Value(rule.GetStatus().GetValue()),
			strategyMetricRuleMutation.Duration.Value(int64(rule.GetDuration())),
		}
		if _, err := strategyMetricRuleMutation.WithContext(ctx).Where(wrapper...).UpdateSimple(mutations...); err != nil {
			return nil, err
		}
		ruleDo, err := strategyMetricRuleMutation.WithContext(ctx).Where(wrapper...).First()
		if err != nil {
			return nil, err
		}
		strategyMetricRuleDos = append(strategyMetricRuleDos, ruleDo)
		ruleDoItem := build.ToStrategyMetricRule(ctx, ruleDo)
		ruleDoItem.WithContext(ctx)
		alarmPages := build.ToDicts(ctx, rule.GetAlarmPages())
		alarmPagesMutation := tx.StrategyMetricRule.AlarmPages.WithContext(ctx).Model(ruleDoItem)
		if len(alarmPages) > 0 {
			if err := alarmPagesMutation.Replace(alarmPages...); err != nil {
				return nil, err
			}
		} else {
			if err := alarmPagesMutation.Clear(); err != nil {
				return nil, err
			}
		}
		noticeGroups := build.ToStrategyNotices(ctx, rule.GetReceiverRoutes())
		noticeGroupsMutation := tx.StrategyMetricRule.Notices.WithContext(ctx).Model(ruleDoItem)
		if len(noticeGroups) > 0 {
			if err := noticeGroupsMutation.Replace(noticeGroups...); err != nil {
				return nil, err
			}
		} else {
			if err := noticeGroupsMutation.Clear(); err != nil {
				return nil, err
			}
		}
		labelNotices := slices.Map(rule.GetLabelNotices(), func(notice bo.LabelNotice) *team.StrategyMetricRuleLabelNotice {
			labelNoticeDo := &team.StrategyMetricRuleLabelNotice{
				LabelKey:             notice.GetKey(),
				LabelValue:           notice.GetValue(),
				Notices:              build.ToStrategyNotices(ctx, notice.GetReceiverRoutes()),
				StrategyMetricRuleID: ruleDoItem.GetID(),
			}
			labelNoticeDo.WithContext(ctx)
			return labelNoticeDo
		})
		strategyMetricRuleLabelNoticeMutation := tx.StrategyMetricRuleLabelNotice
		strategyMetricRuleLabelNoticeWrapper := []gen.Condition{
			strategyMetricRuleLabelNoticeMutation.StrategyMetricRuleID.Eq(ruleDoItem.GetID()),
		}
		if _, err := strategyMetricRuleLabelNoticeMutation.WithContext(ctx).Where(strategyMetricRuleLabelNoticeWrapper...).Delete(); err != nil {
			return nil, err
		}
		if err := strategyMetricRuleLabelNoticeMutation.WithContext(ctx).Create(labelNotices...); err != nil {
			return nil, err
		}
	}
	return strategyMetricRuleDos, nil
}

// CreateLevels implements repository.TeamStrategyMetric.
func (t *teamStrategyMetricImpl) CreateLevels(ctx context.Context, params bo.SaveTeamMetricStrategyLevels) ([]do.StrategyMetricRule, error) {
	tx := getTeamBizQuery(ctx, t)

	strategyMetricRuleDos := slices.Map(params.GetLevels(), func(rule bo.SaveTeamMetricStrategyLevel) *team.StrategyMetricRule {
		ruleItem := &team.StrategyMetricRule{
			StrategyMetricID: params.GetStrategyMetric().GetID(),
			LevelID:          rule.GetLevel().GetID(),
			SampleMode:       rule.GetSampleMode(),
			Total:            rule.GetTotal(),
			Condition:        rule.GetCondition(),
			Values:           rule.GetValues(),
			StrategyMetric:   build.ToStrategyMetric(ctx, params.GetStrategyMetric()),
			Level:            build.ToDict(ctx, rule.GetLevel()),
			Duration:         rule.GetDuration(),
			Status:           rule.GetStatus(),
			Notices:          build.ToStrategyNotices(ctx, rule.GetReceiverRoutes()),
			LabelNotices: slices.Map(rule.GetLabelNotices(), func(notice bo.LabelNotice) *team.StrategyMetricRuleLabelNotice {
				item := &team.StrategyMetricRuleLabelNotice{
					LabelKey:   notice.GetKey(),
					LabelValue: notice.GetValue(),
					Notices:    build.ToStrategyNotices(ctx, notice.GetReceiverRoutes()),
				}
				item.WithContext(ctx)
				return item
			}),
			AlarmPages: build.ToDicts(ctx, rule.GetAlarmPages()),
		}
		ruleItem.WithContext(ctx)
		return ruleItem
	})

	if err := tx.StrategyMetricRule.WithContext(ctx).Create(strategyMetricRuleDos...); err != nil {
		return nil, err
	}

	return slices.Map(strategyMetricRuleDos, func(rule *team.StrategyMetricRule) do.StrategyMetricRule {
		return rule
	}), nil
}

// Delete implements repository.TeamStrategyMetric.
func (t *teamStrategyMetricImpl) Delete(ctx context.Context, params *bo.OperateTeamStrategyParams) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)

	strategyMetricMutation := tx.StrategyMetric
	wrapper := []gen.Condition{
		strategyMetricMutation.StrategyID.Eq(params.StrategyId),
		strategyMetricMutation.TeamID.Eq(teamId),
	}

	_, err := strategyMetricMutation.WithContext(ctx).Where(wrapper...).Delete()
	return err
}

// DeleteLevels implements repository.TeamStrategyMetric.
func (t *teamStrategyMetricImpl) DeleteLevels(ctx context.Context, params *bo.OperateTeamStrategyParams) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)

	strategyMetricMutation := tx.StrategyMetric
	wrapper := []gen.Condition{
		strategyMetricMutation.StrategyID.Eq(params.StrategyId),
		strategyMetricMutation.TeamID.Eq(teamId),
	}
	var strategyMetricIds []uint32
	if err := strategyMetricMutation.WithContext(ctx).Select(strategyMetricMutation.ID).Where(wrapper...).Scan(&strategyMetricIds); err != nil {
		return err
	}
	if len(strategyMetricIds) == 0 {
		return nil
	}
	strategyMetricRuleMutation := tx.StrategyMetricRule
	wrapper = []gen.Condition{
		strategyMetricRuleMutation.TeamID.Eq(teamId),
		strategyMetricRuleMutation.StrategyMetricID.In(strategyMetricIds...),
	}
	_, err := strategyMetricRuleMutation.WithContext(ctx).Where(wrapper...).Delete()
	return err
}

// Get implements repository.TeamStrategyMetric.
func (t *teamStrategyMetricImpl) Get(ctx context.Context, params *bo.OperateTeamStrategyParams) (do.StrategyMetric, error) {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)

	strategyMetricMutation := tx.StrategyMetric
	strategyMetricRuleMutation := tx.StrategyMetricRule
	wrapper := []gen.Condition{
		strategyMetricMutation.StrategyID.Eq(params.StrategyId),
		strategyMetricMutation.TeamID.Eq(teamId),
	}

	preloads := []field.RelationField{
		strategyMetricMutation.Strategy.RelationField,
		strategyMetricMutation.StrategyMetricRules.AlarmPages,
		strategyMetricMutation.StrategyMetricRules.LabelNotices,
		strategyMetricMutation.StrategyMetricRules.LabelNotices.Notices,
		strategyMetricMutation.Datasource,
	}
	if params.StrategyLevelId > 0 {
		preloads = append(preloads, strategyMetricMutation.StrategyMetricRules.Where(strategyMetricRuleMutation.ID.Eq(params.StrategyLevelId)).Level)
	} else {
		preloads = append(preloads, strategyMetricMutation.StrategyMetricRules.Level)
	}
	strategyMetricDo, err := strategyMetricMutation.WithContext(ctx).Preload(preloads...).Where(wrapper...).First()
	if err != nil {
		err = strategyMetricNotFound(err)
		if !merr.IsNotFound(err) {
			return nil, err
		}
		strategyMutation := tx.Strategy
		strategyWrapper := []gen.Condition{
			strategyMutation.ID.Eq(params.StrategyId),
			strategyMutation.TeamID.Eq(teamId),
		}
		strategyDo, err := strategyMutation.WithContext(ctx).Where(strategyWrapper...).First()
		if err != nil {
			return nil, strategyNotFound(err)
		}
		return &team.StrategyMetric{
			Strategy: strategyDo,
		}, nil
	}

	return strategyMetricDo, nil
}
