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

// Create implements repository.TeamStrategyMetricLevel.
func (t *teamStrategyMetricLevelRepoImpl) Create(ctx context.Context, params bo.SaveTeamMetricStrategyLevel) error {
	labelNotices := slices.Map(params.GetLabelNotices(), func(item bo.LabelNotice) *team.StrategyMetricRuleLabelNotice {
		labelNotice := &team.StrategyMetricRuleLabelNotice{
			TeamModel:            do.TeamModel{},
			StrategyMetricRuleID: 0,
			LabelKey:             item.GetKey(),
			LabelValue:           item.GetValue(),
			Notices:              build.ToStrategyNotices(ctx, item.GetReceiverRoutes()),
		}
		labelNotice.WithContext(ctx)
		return labelNotice
	})
	strategyMetricLevel := &team.StrategyMetricRule{
		TeamModel:        do.TeamModel{},
		StrategyMetricID: params.GetStrategyMetricID(),
		LevelID:          0,
		Level:            build.ToDict(ctx, params.GetLevel()),
		SampleMode:       params.GetSampleMode(),
		Condition:        params.GetCondition(),
		Total:            params.GetTotal(),
		Values:           params.GetValues(),
		Duration:         params.GetDuration(),
		Status:           vobj.GlobalStatusEnable,
		Notices:          build.ToStrategyNotices(ctx, params.GetReceiverRoutes()),
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
func (t *teamStrategyMetricLevelRepoImpl) GetByLevelId(ctx context.Context, params *bo.OperateTeamStrategyLevelParams) (do.StrategyMetricRule, error) {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := []gen.Condition{
		tx.StrategyMetricRule.LevelID.Eq(params.StrategyLevelId),
		tx.StrategyMetricRule.StrategyMetricID.Eq(params.StrategyMetricId),
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
func (t *teamStrategyMetricLevelRepoImpl) Update(ctx context.Context, params bo.SaveTeamMetricStrategyLevel) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := []gen.Condition{
		tx.StrategyMetricRule.ID.Eq(params.GetID()),
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
	labelNotices := slices.Map(params.GetLabelNotices(), func(item bo.LabelNotice) *team.StrategyMetricRuleLabelNotice {
		labelNotice := &team.StrategyMetricRuleLabelNotice{
			TeamModel:            do.TeamModel{},
			StrategyMetricRuleID: 0,
			LabelKey:             item.GetKey(),
			LabelValue:           item.GetValue(),
			Notices:              build.ToStrategyNotices(ctx, item.GetReceiverRoutes()),
		}
		labelNotice.WithContext(ctx)
		return labelNotice
	})

	ruleDo, err := t.Get(ctx, params.GetStrategyMetricID())
	if err != nil {
		return err
	}
	rule := build.ToStrategyMetricRule(ctx, ruleDo)
	ruleLabelNoticeMutation := tx.StrategyMetricRule.LabelNotices.WithContext(ctx).Model(rule)
	if len(labelNotices) > 0 {
		ruleLabelNoticeMutation.Replace(labelNotices...)
	} else {
		ruleLabelNoticeMutation.Clear()
	}
	alarmPages := build.ToDicts(ctx, params.GetAlarmPages())
	ruleAlarmPageMutation := tx.StrategyMetricRule.AlarmPages.WithContext(ctx).Model(rule)
	if len(alarmPages) > 0 {
		ruleAlarmPageMutation.Replace(alarmPages...)
	} else {
		ruleAlarmPageMutation.Clear()
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
