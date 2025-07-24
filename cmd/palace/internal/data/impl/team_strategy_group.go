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

func NewTeamStrategyGroupRepo(data *data.Data) repository.TeamStrategyGroup {
	return &teamStrategyGroupRepoImpl{
		Data: data,
	}
}

type teamStrategyGroupRepoImpl struct {
	*data.Data
}

// Create implements repository.TeamStrategyGroup.
func (t *teamStrategyGroupRepoImpl) Create(ctx context.Context, params *bo.SaveTeamStrategyGroupParams) error {
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
func (t *teamStrategyGroupRepoImpl) Delete(ctx context.Context, id uint32) error {
	tx, teamID := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.StrategyGroup
	wrappers := []gen.Condition{
		mutation.ID.Eq(id),
		mutation.TeamID.Eq(teamID),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).Delete()
	if err != nil {
		return err
	}
	strategyMutation := tx.Strategy
	wrappers = []gen.Condition{
		strategyMutation.StrategyGroupID.Eq(id),
		strategyMutation.TeamID.Eq(teamID),
	}
	_, err = strategyMutation.WithContext(ctx).Where(wrappers...).Delete()
	return err
}

// Get implements repository.TeamStrategyGroup.
func (t *teamStrategyGroupRepoImpl) Get(ctx context.Context, id uint32) (do.StrategyGroup, error) {
	tx, teamID := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.StrategyGroup
	wrappers := []gen.Condition{
		mutation.ID.Eq(id),
		mutation.TeamID.Eq(teamID),
	}
	group, err := mutation.WithContext(ctx).Where(wrappers...).Preload(field.Associations).First()
	if err != nil {
		return nil, strategyGroupNotFound(err)
	}
	return group, nil
}

// List implements repository.TeamStrategyGroup.
func (t *teamStrategyGroupRepoImpl) List(ctx context.Context, listParams *bo.ListTeamStrategyGroupParams) (*bo.ListTeamStrategyGroupReply, error) {
	tx, teamID := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.StrategyGroup
	wrappers := mutation.WithContext(ctx).Where(mutation.TeamID.Eq(teamID))
	if validate.TextIsNotNull(listParams.Keyword) {
		wrappers = wrappers.Where(mutation.Name.Like(listParams.Keyword))
	}
	if listParams.Status.Exist() && !listParams.Status.IsUnknown() {
		wrappers = wrappers.Where(mutation.Status.Eq(listParams.Status.GetValue()))
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
	rows := slices.Map(groups, func(group *team.StrategyGroup) do.StrategyGroup { return group })
	return listParams.ToListReply(rows), nil
}

// Select implements repository.TeamStrategyGroup.
func (t *teamStrategyGroupRepoImpl) Select(ctx context.Context, selectParams *bo.SelectTeamStrategyGroupRequest) (*bo.SelectTeamStrategyGroupReply, error) {
	tx, teamID := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.StrategyGroup
	wrappers := mutation.WithContext(ctx).Where(mutation.TeamID.Eq(teamID))
	if validate.TextIsNotNull(selectParams.Keyword) {
		wrappers = wrappers.Where(mutation.Name.Like(selectParams.Keyword))
	}
	if len(selectParams.Status) > 0 {
		wrappers = wrappers.Where(mutation.Status.In(slices.Map(selectParams.Status, func(status vobj.GlobalStatus) int8 { return int8(status) })...))
	}
	if validate.IsNotNil(selectParams.PaginationRequest) {
		total, err := wrappers.Count()
		if err != nil {
			return nil, err
		}
		selectParams.WithTotal(total)
		wrappers = wrappers.Limit(int(selectParams.Limit)).Offset(selectParams.Offset())
	}
	selectColumns := []field.Expr{
		mutation.ID,
		mutation.Name,
		mutation.Remark,
		mutation.Status,
		mutation.DeletedAt,
	}
	groups, err := wrappers.WithContext(ctx).Select(selectColumns...).Order(mutation.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(groups, func(group *team.StrategyGroup) do.StrategyGroup { return group })
	return selectParams.ToSelectReply(rows), nil
}

// Update implements repository.TeamStrategyGroup.
func (t *teamStrategyGroupRepoImpl) Update(ctx context.Context, params *bo.SaveTeamStrategyGroupParams) error {
	tx, teamID := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.StrategyGroup
	wrappers := []gen.Condition{
		mutation.ID.Eq(params.ID),
		mutation.TeamID.Eq(teamID),
	}
	mutations := []field.AssignExpr{
		mutation.Name.Value(params.Name),
		mutation.Remark.Value(params.Remark),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).UpdateSimple(mutations...)
	return err
}

// UpdateStatus implements repository.TeamStrategyGroup.
func (t *teamStrategyGroupRepoImpl) UpdateStatus(ctx context.Context, params *bo.UpdateTeamStrategyGroupStatusParams) error {
	tx, teamID := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.StrategyGroup
	wrappers := []gen.Condition{
		mutation.ID.Eq(params.ID),
		mutation.TeamID.Eq(teamID),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).
		UpdateSimple(mutation.Status.Value(params.Status.GetValue()))
	return err
}
