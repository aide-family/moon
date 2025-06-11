package impl

import (
	"context"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/team"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/cmd/palace/internal/data"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func NewTeamStrategyGroupRepo(data *data.Data) repository.TeamStrategyGroup {
	return &teamStrategyGroupRepo{
		Data: data,
	}
}

type teamStrategyGroupRepo struct {
	*data.Data
}

// Create implements repository.TeamStrategyGroup.
func (t *teamStrategyGroupRepo) Create(ctx context.Context, params *bo.SaveTeamStrategyGroupParams) error {
	tx := getTeamBizQuery(ctx, t)
	groupDo := &team.StrategyGroup{
		Name:   params.Name,
		Remark: params.Remark,
		Status: vobj.GlobalStatusEnable,
	}
	groupDo.WithContext(ctx)

	return tx.StrategyGroup.WithContext(ctx).Create(groupDo)
}

// Delete implements repository.TeamStrategyGroup.
func (t *teamStrategyGroupRepo) Delete(ctx context.Context, id uint32) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.StrategyGroup
	wrappers := []gen.Condition{
		mutation.ID.Eq(id),
		mutation.TeamID.Eq(teamId),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).Delete()
	if err != nil {
		return err
	}
	strategyMutation := tx.Strategy
	wrappers = []gen.Condition{
		strategyMutation.StrategyGroupID.Eq(id),
		strategyMutation.TeamID.Eq(teamId),
	}
	_, err = strategyMutation.WithContext(ctx).Where(wrappers...).Delete()
	return err
}

// Get implements repository.TeamStrategyGroup.
func (t *teamStrategyGroupRepo) Get(ctx context.Context, id uint32) (do.StrategyGroup, error) {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.StrategyGroup
	wrappers := []gen.Condition{
		mutation.ID.Eq(id),
		mutation.TeamID.Eq(teamId),
	}
	group, err := mutation.WithContext(ctx).Preload(field.Associations).Where(wrappers...).First()
	if err != nil {
		return nil, strategyGroupNotFound(err)
	}
	return group, nil
}

// List implements repository.TeamStrategyGroup.
func (t *teamStrategyGroupRepo) List(ctx context.Context, listParams *bo.ListTeamStrategyGroupParams) (*bo.ListTeamStrategyGroupReply, error) {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.StrategyGroup
	wrappers := mutation.WithContext(ctx).Where(mutation.TeamID.Eq(teamId))
	if validate.TextIsNotNull(listParams.Keyword) {
		wrappers = wrappers.Where(mutation.Name.Like(listParams.Keyword))
	}
	if len(listParams.Status) > 0 {
		wrappers = wrappers.Where(mutation.Status.In(slices.Map(listParams.Status, func(status vobj.GlobalStatus) int8 { return int8(status) })...))
	}
	if validate.IsNotNil(listParams.PaginationRequest) {
		total, err := wrappers.Count()
		if err != nil {
			return nil, err
		}
		listParams.WithTotal(total)
		wrappers = wrappers.Limit(int(listParams.Limit)).Offset(listParams.Offset())
	}
	groups, err := wrappers.Find()
	if err != nil {
		return nil, err
	}
	return listParams.ToListTeamStrategyGroupReply(groups), nil
}

// Update implements repository.TeamStrategyGroup.
func (t *teamStrategyGroupRepo) Update(ctx context.Context, params *bo.SaveTeamStrategyGroupParams) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.StrategyGroup
	wrappers := []gen.Condition{
		mutation.ID.Eq(params.ID),
		mutation.TeamID.Eq(teamId),
	}
	mutations := []field.AssignExpr{
		mutation.Name.Value(params.Name),
		mutation.Remark.Value(params.Remark),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).UpdateSimple(mutations...)
	return err
}

// UpdateStatus implements repository.TeamStrategyGroup.
func (t *teamStrategyGroupRepo) UpdateStatus(ctx context.Context, params *bo.UpdateTeamStrategyGroupStatusParams) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.StrategyGroup
	wrappers := []gen.Condition{
		mutation.ID.Eq(params.ID),
		mutation.TeamID.Eq(teamId),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).
		UpdateSimple(mutation.Status.Value(params.Status.GetValue()))
	return err
}
