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
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

// NewTimeEngineRule 创建时间引擎规则仓储实现
func NewTimeEngineRule(data *data.Data) repository.TimeEngineRule {
	return &timeEngineRuleRepoImpl{
		Data: data,
	}
}

type timeEngineRuleRepoImpl struct {
	*data.Data
}

// CreateTimeEngineRule 创建时间引擎规则
func (r *timeEngineRuleRepoImpl) CreateTimeEngineRule(ctx context.Context, req *bo.SaveTimeEngineRuleRequest) error {
	timeEngineRule := &team.TimeEngineRule{
		Name:   req.Name,
		Remark: req.Remark,
		Type:   req.Type,
		Status: vobj.GlobalStatusEnable,
		Rule:   req.Rules,
	}
	timeEngineRule.WithContext(ctx)
	bizQuery := getTeamBizQuery(ctx, r)
	return bizQuery.TimeEngineRule.WithContext(ctx).Create(timeEngineRule)
}

// UpdateTimeEngineRule 更新时间引擎规则
func (r *timeEngineRuleRepoImpl) UpdateTimeEngineRule(ctx context.Context, timeEngineRuleId uint32, req *bo.SaveTimeEngineRuleRequest) error {
	bizMutation, teamId := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineRuleMutation := bizMutation.TimeEngineRule
	wrappers := []gen.Condition{
		timeEngineRuleMutation.ID.Eq(timeEngineRuleId),
		timeEngineRuleMutation.TeamID.Eq(teamId),
	}
	columns := []field.AssignExpr{
		timeEngineRuleMutation.Name.Value(req.Name),
		timeEngineRuleMutation.Remark.Value(req.Remark),
		timeEngineRuleMutation.Type.Value(req.Type.GetValue()),
		timeEngineRuleMutation.Rule.Value(team.Rules(req.Rules)),
	}

	_, err := timeEngineRuleMutation.WithContext(ctx).Where(wrappers...).UpdateSimple(columns...)
	return err
}

// UpdateTimeEngineRuleStatus 更新时间引擎规则状态
func (r *timeEngineRuleRepoImpl) UpdateTimeEngineRuleStatus(ctx context.Context, req *bo.UpdateTimeEngineRuleStatusRequest) error {
	if len(req.TimeEngineRuleIds) == 0 {
		return nil
	}
	bizMutation, teamId := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineRuleMutation := bizMutation.TimeEngineRule
	wrappers := []gen.Condition{
		timeEngineRuleMutation.ID.In(req.TimeEngineRuleIds...),
		timeEngineRuleMutation.TeamID.Eq(teamId),
	}
	columns := []field.AssignExpr{
		timeEngineRuleMutation.Status.Value(req.Status.GetValue()),
	}
	_, err := timeEngineRuleMutation.WithContext(ctx).Where(wrappers...).UpdateSimple(columns...)
	return err
}

// DeleteTimeEngineRule 删除时间引擎规则
func (r *timeEngineRuleRepoImpl) DeleteTimeEngineRule(ctx context.Context, req *bo.DeleteTimeEngineRuleRequest) error {
	bizMutation, teamId := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineRuleMutation := bizMutation.TimeEngineRule
	wrappers := []gen.Condition{
		timeEngineRuleMutation.ID.Eq(req.TimeEngineRuleId),
		timeEngineRuleMutation.TeamID.Eq(teamId),
	}
	_, err := timeEngineRuleMutation.WithContext(ctx).Where(wrappers...).Delete()
	return err
}

// GetTimeEngineRule 获取时间引擎规则详情
func (r *timeEngineRuleRepoImpl) GetTimeEngineRule(ctx context.Context, req *bo.GetTimeEngineRuleRequest) (do.TimeEngineRule, error) {
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineRuleQuery := bizQuery.TimeEngineRule
	timeEngineRule, err := timeEngineRuleQuery.WithContext(ctx).
		Where(timeEngineRuleQuery.ID.Eq(req.TimeEngineRuleId), timeEngineRuleQuery.TeamID.Eq(teamId)).
		First()
	if err != nil {
		return nil, teamTimeEngineRuleNotFound(err)
	}

	return timeEngineRule, nil
}

// ListTimeEngineRule 获取时间引擎规则列表
func (r *timeEngineRuleRepoImpl) ListTimeEngineRule(ctx context.Context, req *bo.ListTimeEngineRuleRequest) (*bo.ListTimeEngineRuleReply, error) {
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineRuleQuery := bizQuery.TimeEngineRule
	timeEngineRuleWrapper := timeEngineRuleQuery.Where(timeEngineRuleQuery.TeamID.Eq(teamId))

	if !req.Status.IsUnknown() {
		timeEngineRuleWrapper = timeEngineRuleWrapper.Where(timeEngineRuleQuery.Status.Eq(req.Status.GetValue()))
	}

	if validate.TextIsNotNull(req.Keyword) {
		or := []gen.Condition{
			timeEngineRuleQuery.Name.Like(req.Keyword),
			timeEngineRuleQuery.Remark.Like(req.Keyword),
		}
		timeEngineRuleWrapper = timeEngineRuleWrapper.Where(timeEngineRuleQuery.Or(or...))
	}

	if len(req.Types) > 0 {
		types := slices.MapFilter(req.Types, func(v vobj.TimeEngineRuleType) (int8, bool) {
			if !v.Exist() || v.IsUnknown() {
				return 0, false
			}
			return v.GetValue(), true
		})
		timeEngineRuleWrapper = timeEngineRuleWrapper.Where(timeEngineRuleQuery.Type.In(types...))
	}

	if validate.IsNotNil(req.PaginationRequest) {
		total, err := timeEngineRuleWrapper.WithContext(ctx).Count()
		if err != nil {
			return nil, err
		}
		req.WithTotal(total)
		timeEngineRuleWrapper = timeEngineRuleWrapper.Limit(int(req.Limit)).Offset(int(req.Offset()))
	}
	timeEngineRuleWrapper = timeEngineRuleWrapper.Order(timeEngineRuleQuery.CreatedAt.Desc())
	timeEngineRules, err := timeEngineRuleWrapper.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}

	dos := slices.Map(timeEngineRules, func(v *team.TimeEngineRule) do.TimeEngineRule { return v })
	return req.ToListReply(dos), nil
}

// SelectTimeEngineRule 获取时间引擎规则列表
func (r *timeEngineRuleRepoImpl) SelectTimeEngineRule(ctx context.Context, req *bo.SelectTimeEngineRuleRequest) (*bo.SelectTimeEngineRuleReply, error) {
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineRuleQuery := bizQuery.TimeEngineRule
	timeEngineRuleWrapper := timeEngineRuleQuery.Where(timeEngineRuleQuery.TeamID.Eq(teamId))

	if !req.Status.IsUnknown() {
		timeEngineRuleWrapper = timeEngineRuleWrapper.Where(timeEngineRuleQuery.Status.Eq(req.Status.GetValue()))
	}

	if validate.IsNotNil(req.PaginationRequest) {
		total, err := timeEngineRuleWrapper.WithContext(ctx).Count()
		if err != nil {
			return nil, err
		}
		timeEngineRuleWrapper = timeEngineRuleWrapper.Limit(int(req.Limit)).Offset(int(req.Offset()))
		req.WithTotal(total)
	}
	selectColumns := []field.Expr{
		timeEngineRuleQuery.ID,
		timeEngineRuleQuery.Name,
		timeEngineRuleQuery.Remark,
		timeEngineRuleQuery.Status,
		timeEngineRuleQuery.DeletedAt,
		timeEngineRuleQuery.Type,
	}
	timeEngineRules, err := timeEngineRuleWrapper.WithContext(ctx).Select(selectColumns...).Order(timeEngineRuleQuery.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(timeEngineRules, func(v *team.TimeEngineRule) do.TimeEngineRule { return v })
	return req.ToSelectReply(rows), nil
}

// Find 获取时间引擎规则列表
func (r *timeEngineRuleRepoImpl) Find(ctx context.Context, ruleIds ...uint32) ([]do.TimeEngineRule, error) {
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineRuleQuery := bizQuery.TimeEngineRule
	wrappers := timeEngineRuleQuery.Where(timeEngineRuleQuery.TeamID.Eq(teamId))
	if len(ruleIds) > 0 {
		wrappers = wrappers.Where(timeEngineRuleQuery.ID.In(ruleIds...))
	}
	timeEngineRules, err := wrappers.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}
	dos := slices.Map(timeEngineRules, func(v *team.TimeEngineRule) do.TimeEngineRule { return v })
	return dos, nil
}
