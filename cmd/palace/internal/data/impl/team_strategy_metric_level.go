package impl

import (
	"context"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/cmd/palace/internal/data/impl/build"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
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
	strategyMetricLevel := &team.StrategyMetricRule{
		StrategyMetricID: params.GetStrategyMetric().GetID(),
		LevelID:          params.GetLevel().GetID(),
		SampleMode:       params.GetSampleMode(),
		Condition:        params.GetCondition(),
		Total:            params.GetTotal(),
		Values:           params.GetValues(),
		Duration:         params.GetDuration(),
		Status:           vobj.GlobalStatusEnable,
		Notices:          build.ToTeamNoticeGroups(ctx, params.GetNoticeGroupDos()),
		AlarmPages:       build.ToDicts(ctx, params.GetAlarmPages()),
		StrategyID:       params.GetStrategyMetric().GetStrategyID(),
	}
	strategyMetricLevel.WithContext(ctx)
	tx := getTeamBizQuery(ctx, t)
	if err := tx.WithContext(ctx).StrategyMetricRule.Create(strategyMetricLevel); err != nil {
		return err
	}
	labelNotices := slices.Map(params.GetLabelNotices(), func(item *bo.LabelNoticeParams) *team.StrategyMetricRuleLabelNotice {
		labelNotice := &team.StrategyMetricRuleLabelNotice{
			StrategyMetricRuleID: strategyMetricLevel.ID,
			LabelKey:             item.Key,
			LabelValue:           item.Value,
			Notices:              build.ToTeamNoticeGroups(ctx, item.GetNoticeGroupDos()),
		}
		labelNotice.WithContext(ctx)
		return labelNotice
	})
	return tx.WithContext(ctx).StrategyMetricRuleLabelNotice.Create(labelNotices...)
}

// Delete implements repository.TeamStrategyMetricLevel.
func (t *teamStrategyMetricLevelRepoImpl) Delete(ctx context.Context, strategyMetricLevelIds []uint32) error {
	if len(strategyMetricLevelIds) == 0 {
		return nil
	}
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := []gen.Condition{
		tx.StrategyMetricRule.ID.In(strategyMetricLevelIds...),
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
	strategyMetricRuleDo, err := tx.StrategyMetricRule.WithContext(ctx).Where(wrapper...).Preload(field.Associations).First()
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
	strategyMetricRuleDo, err := tx.StrategyMetricRule.WithContext(ctx).Where(wrapper...).Preload(field.Associations).First()
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
	wrapper = wrapper.Preload(field.Associations)

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
		ruleMutation.StrategyID.Value(params.GetStrategyMetric().GetStrategyID()),
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
			Notices:              build.ToTeamNoticeGroups(ctx, item.GetNoticeGroupDos()),
		}
		labelNotice.WithContext(ctx)
		return labelNotice
	})

	ruleDo, err := t.Get(ctx, params.GetStrategyMetricLevel().GetID())
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
	if len(params.StrategyMetricLevelIds) == 0 {
		return nil
	}
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	ruleMutation := tx.StrategyMetricRule
	wrapper := []gen.Condition{
		ruleMutation.ID.In(params.StrategyMetricLevelIds...),
		ruleMutation.TeamID.Eq(teamId),
	}
	_, err := ruleMutation.WithContext(ctx).Where(wrapper...).UpdateSimple(ruleMutation.Status.Value(params.Status.GetValue()))
	return err
}

func (t *teamStrategyMetricLevelRepoImpl) FindByIds(ctx context.Context, strategyMetricLevelIds []uint32) ([]do.StrategyMetricRule, error) {
	if len(strategyMetricLevelIds) == 0 {
		return nil, nil
	}
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := []gen.Condition{
		tx.StrategyMetricRule.ID.In(strategyMetricLevelIds...),
		tx.StrategyMetricRule.TeamID.Eq(teamId),
	}
	rows, err := tx.StrategyMetricRule.WithContext(ctx).Where(wrapper...).Preload(field.Associations).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(rows, func(item *team.StrategyMetricRule) do.StrategyMetricRule {
		return item
	}), nil
}
