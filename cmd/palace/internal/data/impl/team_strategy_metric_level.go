package impl

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/cmd/palace/internal/data/impl/build"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

func NewTeamStrategyMetricLevelRepo(d *data.Data) repository.TeamStrategyMetricLevel {
	return &teamStrategyMetricLevelRepoImpl{
		Data: d,
	}
}

type teamStrategyMetricLevelRepoImpl struct {
	*data.Data
}

// DeleteByStrategyIds implements repository.TeamStrategyMetricLevel.
func (t *teamStrategyMetricLevelRepoImpl) DeleteByStrategyIds(ctx context.Context, strategyIds ...uint32) error {
	if len(strategyIds) == 0 {
		return nil
	}
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.StrategyMetricRule
	wrappers := []gen.Condition{
		mutation.TeamID.Eq(teamId),
		mutation.StrategyID.In(strategyIds...),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).Delete()
	return err
}

// Create implements repository.TeamStrategyMetricLevel.
func (t *teamStrategyMetricLevelRepoImpl) Create(ctx context.Context, params bo.CreateTeamMetricStrategyLevelParams) error {
	labelNotices := slices.Map(params.GetLabelNotices(), func(item *bo.LabelNoticeParams) *team.StrategyMetricRuleLabelNotice {
		labelNotice := &team.StrategyMetricRuleLabelNotice{
			TeamModel:            do.TeamModel{},
			StrategyMetricRuleID: 0,
			LabelKey:             item.Key,
			LabelValue:           item.Value,
			Notices:              build.ToStrategyNotices(ctx, item.GetNoticeGroupDos()),
		}
		labelNotice.WithContext(ctx)
		return labelNotice
	})
	strategyMetricLevel := &team.StrategyMetricRule{
		TeamModel:        do.TeamModel{},
		StrategyMetricID: params.GetStrategyMetric().GetID(),
		LevelID:          0,
		Level:            build.ToDict(ctx, params.GetLevel()),
		SampleMode:       params.GetSampleMode(),
		Condition:        params.GetCondition(),
		Total:            params.GetTotal(),
		Values:           params.GetValues(),
		Duration:         params.GetDuration(),
		Status:           vobj.GlobalStatusEnable,
		Notices:          build.ToStrategyNotices(ctx, params.GetNoticeGroupDos()),
		LabelNotices:     labelNotices,
		AlarmPages:       build.ToDicts(ctx, params.GetAlarmPages()),
	}
	strategyMetricLevel.WithContext(ctx)
	tx := getTeamBizQuery(ctx, t)
	return tx.WithContext(ctx).StrategyMetricRule.Create(strategyMetricLevel)
}

// Delete implements repository.TeamStrategyMetricLevel.
func (t *teamStrategyMetricLevelRepoImpl) Delete(ctx context.Context, strategyMetricLevelId uint32) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := []gen.Condition{
		tx.StrategyMetricRule.ID.Eq(strategyMetricLevelId),
		tx.StrategyMetricRule.TeamID.Eq(teamId),
	}
	_, err := tx.WithContext(ctx).StrategyMetricRule.Where(wrapper...).Delete()
	return err
}

// DeleteByStrategyId implements repository.TeamStrategyMetricLevel.
func (t *teamStrategyMetricLevelRepoImpl) DeleteByStrategyId(ctx context.Context, strategyId uint32) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := []gen.Condition{
		tx.StrategyMetricRule.StrategyID.Eq(strategyId),
		tx.StrategyMetricRule.TeamID.Eq(teamId),
	}
	_, err := tx.WithContext(ctx).StrategyMetricRule.Where(wrapper...).Delete()
	return err
}

// Get implements repository.TeamStrategyMetricLevel.
func (t *teamStrategyMetricLevelRepoImpl) Get(ctx context.Context, strategyMetricLevelId uint32) (do.StrategyMetricRule, error) {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := []gen.Condition{
		tx.StrategyMetricRule.ID.Eq(strategyMetricLevelId),
		tx.StrategyMetricRule.TeamID.Eq(teamId),
	}
	strategyMetricRuleDo, err := tx.StrategyMetricRule.WithContext(ctx).Where(wrapper...).First()
	if err != nil {
		return nil, strategyMetricRuleNotFound(err)
	}
	return strategyMetricRuleDo, nil
}

// GetByLevelId implements repository.TeamStrategyMetricLevel.
func (t *teamStrategyMetricLevelRepoImpl) GetByLevelId(ctx context.Context, strategyMetricId uint32, levelId uint32) (do.StrategyMetricRule, error) {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := []gen.Condition{
		tx.StrategyMetricRule.LevelID.Eq(levelId),
		tx.StrategyMetricRule.StrategyMetricID.Eq(strategyMetricId),
		tx.StrategyMetricRule.TeamID.Eq(teamId),
	}
	strategyMetricRuleDo, err := tx.StrategyMetricRule.WithContext(ctx).Where(wrapper...).First()
	if err != nil {
		return nil, strategyMetricRuleNotFound(err)
	}
	return strategyMetricRuleDo, nil
}

// List implements repository.TeamStrategyMetricLevel.
func (t *teamStrategyMetricLevelRepoImpl) List(ctx context.Context, params *bo.ListTeamMetricStrategyLevelsParams) (*bo.ListTeamMetricStrategyLevelsReply, error) {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	ruleQuery := tx.StrategyMetricRule
	wrapper := ruleQuery.Where(ruleQuery.StrategyMetricID.Eq(params.StrategyMetricID))
	wrapper = wrapper.Where(ruleQuery.TeamID.Eq(teamId))

	if params.LevelId > 0 {
		wrapper = wrapper.Where(ruleQuery.LevelID.Eq(params.LevelId))
	}
	if validate.IsNotNil(params.PaginationRequest) {
		total, err := wrapper.WithContext(ctx).Count()
		if err != nil {
			return nil, err
		}
		params.WithTotal(total)
		wrapper = wrapper.Limit(int(params.Limit)).Offset(params.Offset())
	}

	strategyMetricRuleDos, err := wrapper.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(strategyMetricRuleDos, func(item *team.StrategyMetricRule) do.StrategyMetricRule {
		return item
	})
	return params.ToListReply(rows), nil
}

// Update implements repository.TeamStrategyMetricLevel.
func (t *teamStrategyMetricLevelRepoImpl) Update(ctx context.Context, params bo.UpdateTeamMetricStrategyLevelParams) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := []gen.Condition{
		tx.StrategyMetricRule.ID.Eq(params.GetStrategyMetricLevel().GetID()),
		tx.StrategyMetricRule.TeamID.Eq(teamId),
	}
	ruleMutation := tx.StrategyMetricRule

	mutations := []field.AssignExpr{
		ruleMutation.LevelID.Value(params.GetLevel().GetID()),
		ruleMutation.SampleMode.Value(params.GetSampleMode().GetValue()),
		ruleMutation.Condition.Value(params.GetCondition().GetValue()),
		ruleMutation.Total.Value(params.GetTotal()),
		ruleMutation.Values.Value(team.Values(params.GetValues())),
		ruleMutation.Duration.Value(params.GetDuration()),
	}
	_, err := tx.StrategyMetricRule.WithContext(ctx).Where(wrapper...).UpdateSimple(mutations...)
	if err != nil {
		return err
	}
	labelNotices := slices.Map(params.GetLabelNotices(), func(item *bo.LabelNoticeParams) *team.StrategyMetricRuleLabelNotice {
		labelNotice := &team.StrategyMetricRuleLabelNotice{
			TeamModel:            do.TeamModel{},
			StrategyMetricRuleID: 0,
			LabelKey:             item.Key,
			LabelValue:           item.Value,
			Notices:              build.ToStrategyNotices(ctx, item.GetNoticeGroupDos()),
		}
		labelNotice.WithContext(ctx)
		return labelNotice
	})

	ruleDo, err := t.Get(ctx, params.GetStrategyMetricLevel().GetStrategyMetricID())
	if err != nil {
		return err
	}
	rule := build.ToStrategyMetricRule(ctx, ruleDo)
	ruleLabelNoticeMutation := tx.StrategyMetricRule.LabelNotices.WithContext(ctx).Model(rule)
	if len(labelNotices) > 0 {
		if err := ruleLabelNoticeMutation.Replace(labelNotices...); err != nil {
			return err
		}
	} else {
		if err := ruleLabelNoticeMutation.Clear(); err != nil {
			return err
		}
	}
	alarmPages := build.ToDicts(ctx, params.GetAlarmPages())
	ruleAlarmPageMutation := tx.StrategyMetricRule.AlarmPages.WithContext(ctx).Model(rule)
	if len(alarmPages) > 0 {
		if err := ruleAlarmPageMutation.Replace(alarmPages...); err != nil {
			return err
		}
	} else {
		if err := ruleAlarmPageMutation.Clear(); err != nil {
			return err
		}
	}

	return nil
}

// UpdateStatus implements repository.TeamStrategyMetricLevel.
func (t *teamStrategyMetricLevelRepoImpl) UpdateStatus(ctx context.Context, params *bo.UpdateTeamMetricStrategyLevelStatusParams) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	ruleMutation := tx.StrategyMetricRule
	wrapper := []gen.Condition{
		ruleMutation.ID.Eq(params.StrategyMetricLevelID),
		ruleMutation.TeamID.Eq(teamId),
	}
	_, err := ruleMutation.WithContext(ctx).Where(wrapper...).UpdateSimple(ruleMutation.Status.Value(params.Status.GetValue()))
	return err
}
