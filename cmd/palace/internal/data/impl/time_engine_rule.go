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

// NewTimeEngineRule creates a time engine rule repository implementation
func NewTimeEngineRule(data *data.Data) repository.TimeEngineRule {
	return &timeEngineRuleRepoImpl{
		Data: data,
	}
}

type timeEngineRuleRepoImpl struct {
	*data.Data
}

// CreateTimeEngineRule creates a time engine rule
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

// UpdateTimeEngineRule updates a time engine rule
func (r *timeEngineRuleRepoImpl) UpdateTimeEngineRule(ctx context.Context, timeEngineRuleID uint32, req *bo.SaveTimeEngineRuleRequest) error {
	bizMutation, teamID := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineRuleMutation := bizMutation.TimeEngineRule
	wrappers := []gen.Condition{
		timeEngineRuleMutation.ID.Eq(timeEngineRuleID),
		timeEngineRuleMutation.TeamID.Eq(teamID),
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

// UpdateTimeEngineRuleStatus updates the status of a time engine rule
func (r *timeEngineRuleRepoImpl) UpdateTimeEngineRuleStatus(ctx context.Context, req *bo.UpdateTimeEngineRuleStatusRequest) error {
	if len(req.TimeEngineRuleIds) == 0 {
		return nil
	}
	bizMutation, teamID := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineRuleMutation := bizMutation.TimeEngineRule
	wrappers := []gen.Condition{
		timeEngineRuleMutation.ID.In(req.TimeEngineRuleIds...),
		timeEngineRuleMutation.TeamID.Eq(teamID),
	}
	columns := []field.AssignExpr{
		timeEngineRuleMutation.Status.Value(req.Status.GetValue()),
	}
	_, err := timeEngineRuleMutation.WithContext(ctx).Where(wrappers...).UpdateSimple(columns...)
	return err
}

// DeleteTimeEngineRule deletes a time engine rule
func (r *timeEngineRuleRepoImpl) DeleteTimeEngineRule(ctx context.Context, req *bo.DeleteTimeEngineRuleRequest) error {
	bizMutation, teamID := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineRuleMutation := bizMutation.TimeEngineRule
	wrappers := []gen.Condition{
		timeEngineRuleMutation.ID.Eq(req.TimeEngineRuleID),
		timeEngineRuleMutation.TeamID.Eq(teamID),
	}
	_, err := timeEngineRuleMutation.WithContext(ctx).Where(wrappers...).Delete()
	return err
}

// GetTimeEngineRule gets the details of a time engine rule
func (r *timeEngineRuleRepoImpl) GetTimeEngineRule(ctx context.Context, req *bo.GetTimeEngineRuleRequest) (do.TimeEngineRule, error) {
	bizQuery, teamID := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineRuleQuery := bizQuery.TimeEngineRule
	timeEngineRule, err := timeEngineRuleQuery.WithContext(ctx).
		Where(timeEngineRuleQuery.ID.Eq(req.TimeEngineRuleID), timeEngineRuleQuery.TeamID.Eq(teamID)).
		First()
	if err != nil {
		return nil, teamTimeEngineRuleNotFound(err)
	}

	return timeEngineRule, nil
}

// ListTimeEngineRule gets the list of time engine rules
func (r *timeEngineRuleRepoImpl) ListTimeEngineRule(ctx context.Context, req *bo.ListTimeEngineRuleRequest) (*bo.ListTimeEngineRuleReply, error) {
	bizQuery, teamID := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineRuleQuery := bizQuery.TimeEngineRule
	timeEngineRuleWrapper := timeEngineRuleQuery.Where(timeEngineRuleQuery.TeamID.Eq(teamID))

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

// SelectTimeEngineRule gets the list of time engine rules
func (r *timeEngineRuleRepoImpl) SelectTimeEngineRule(ctx context.Context, req *bo.SelectTimeEngineRuleRequest) (*bo.SelectTimeEngineRuleReply, error) {
	bizQuery, teamID := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineRuleQuery := bizQuery.TimeEngineRule
	timeEngineRuleWrapper := timeEngineRuleQuery.Where(timeEngineRuleQuery.TeamID.Eq(teamID))

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

// Find gets the list of time engine rules
func (r *timeEngineRuleRepoImpl) Find(ctx context.Context, ruleIds ...uint32) ([]do.TimeEngineRule, error) {
	bizQuery, teamID := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineRuleQuery := bizQuery.TimeEngineRule
	wrappers := timeEngineRuleQuery.Where(timeEngineRuleQuery.TeamID.Eq(teamID))
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
