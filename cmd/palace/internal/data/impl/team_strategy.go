package impl

import (
	"context"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/cmd/palace/internal/data/impl/build"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewTeamStrategyRepo(data *data.Data) repository.TeamStrategy {
	return &teamStrategyRepoImpl{
		Data: data,
	}
}

type teamStrategyRepoImpl struct {
	*data.Data
}

// DeleteByStrategyIds implements repository.TeamStrategy.
func (t *teamStrategyRepoImpl) DeleteByStrategyIds(ctx context.Context, strategyIds ...uint32) error {
	if len(strategyIds) == 0 {
		return nil
	}
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.Strategy
	wrappers := []gen.Condition{
		mutation.TeamID.Eq(teamId),
		mutation.ID.In(strategyIds...),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).Delete()
	return err
}

// FindByStrategiesGroupId implements repository.TeamStrategy.
func (t *teamStrategyRepoImpl) FindByStrategiesGroupId(ctx context.Context, strategyGroupId uint32) ([]do.Strategy, error) {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	query := tx.Strategy
	wrapper := query.WithContext(ctx).Where(query.TeamID.Eq(teamId), query.StrategyGroupID.Eq(strategyGroupId))
	rows, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(rows, func(row *team.Strategy) do.Strategy { return row }), nil
}

// GetByName implements repository.TeamStrategy.
func (t *teamStrategyRepoImpl) GetByName(ctx context.Context, name string) (do.Strategy, error) {
	tx := getTeamBizQuery(ctx, t)
	strategy, err := tx.Strategy.WithContext(ctx).Where(tx.Strategy.Name.Eq(name)).First()
	if err != nil {
		return nil, strategyNotFound(err)
	}
	return strategy, nil
}

// NameExists implements repository.TeamStrategy.
func (t *teamStrategyRepoImpl) NameExists(ctx context.Context, name string, strategyId uint32) error {
	tx := getTeamBizQuery(ctx, t)
	_, err := tx.Strategy.WithContext(ctx).Where(tx.Strategy.Name.Eq(name), tx.Strategy.ID.Neq(strategyId)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	return merr.ErrorExist("strategy name %s already exists", name)
}

// Create implements repository.TeamStrategy.
func (t *teamStrategyRepoImpl) Create(ctx context.Context, params bo.CreateTeamStrategyParams) error {
	strategyDo := &team.Strategy{
		StrategyGroupID: params.GetStrategyGroup().GetID(),
		Name:            params.GetName(),
		Remark:          params.GetRemark(),
		Status:          vobj.GlobalStatusDisable,
		StrategyType:    params.GetStrategyType(),
	}
	strategyDo.WithContext(ctx)
	tx := getTeamBizQuery(ctx, t)
	if err := tx.Strategy.WithContext(ctx).Create(strategyDo); err != nil {
		return err
	}
	notices := build.ToTeamNoticeGroups(ctx, params.GetReceiverRoutes())
	if len(notices) > 0 {
		notice := tx.Strategy.Notices.WithContext(ctx).Model(strategyDo)
		if err := notice.Append(notices...); err != nil {
			return err
		}
	}
	return nil
}

// Delete implements repository.TeamStrategy.
func (t *teamStrategyRepoImpl) Delete(ctx context.Context, strategyId uint32) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.Strategy
	wrappers := []gen.Condition{
		mutation.ID.Eq(strategyId),
		mutation.TeamID.Eq(teamId),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).Delete()
	return err
}

// Subscribe implements repository.TeamStrategy.
func (t *teamStrategyRepoImpl) Subscribe(ctx context.Context, params *bo.SubscribeTeamStrategyParams) error {
	tx := getTeamBizQuery(ctx, t)
	subscriberDo := &team.StrategySubscriber{
		StrategyID:    params.StrategyId,
		SubscribeType: params.NoticeType,
	}
	subscriberDo.WithContext(ctx)

	if err := tx.StrategySubscriber.WithContext(ctx).Create(subscriberDo); err != nil {
		return err
	}
	return nil
}

// SubscribeList implements repository.TeamStrategy.
func (t *teamStrategyRepoImpl) SubscribeList(ctx context.Context, params *bo.SubscribeTeamStrategiesParams) (*bo.SubscribeTeamStrategiesReply, error) {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	query := tx.StrategySubscriber
	wrappers := query.WithContext(ctx).Where(query.TeamID.Eq(teamId))
	if len(params.Subscribers) > 0 {
		wrappers = wrappers.Where(query.CreatorID.In(params.Subscribers...))
	}
	if !params.NoticeType.IsUnknown() {
		wrappers = wrappers.Where(query.SubscribeType.Eq(params.NoticeType.GetValue()))
	}
	if validate.IsNotNil(params.PaginationRequest) {
		total, err := wrappers.Count()
		if err != nil {
			return nil, err
		}
		params.WithTotal(total)
		wrappers = wrappers.Limit(int(params.Limit)).Offset(params.Offset())
	}
	subscribers, err := wrappers.Preload(field.Associations).Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(subscribers, func(subscriber *team.StrategySubscriber) do.TeamStrategySubscriber { return subscriber })
	return params.ToListReply(rows), nil
}

// Update implements repository.TeamStrategy.
func (t *teamStrategyRepoImpl) Update(ctx context.Context, params bo.UpdateTeamStrategyParams) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.Strategy
	wrappers := []gen.Condition{
		mutation.ID.Eq(params.GetStrategy().GetID()),
		mutation.TeamID.Eq(teamId),
	}
	mutations := []field.AssignExpr{
		mutation.Name.Value(params.GetName()),
		mutation.Remark.Value(params.GetRemark()),
		mutation.StrategyType.Value(params.GetStrategyType().GetValue()),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).UpdateSimple(mutations...)
	if err != nil {
		return err
	}
	notices := build.ToTeamNoticeGroups(ctx, params.GetReceiverRoutes())
	strategyDo := &team.Strategy{TeamModel: build.ToTeamModel(ctx, params.GetStrategy())}
	notice := tx.Strategy.Notices.WithContext(ctx).Model(strategyDo)
	if len(notices) > 0 {
		if err := notice.Replace(notices...); err != nil {
			return err
		}
	} else {
		if err := notice.Clear(); err != nil {
			return err
		}
	}

	return nil
}

// List implements repository.TeamStrategy.
func (t *teamStrategyRepoImpl) List(ctx context.Context, params *bo.ListTeamStrategyParams) (*bo.ListTeamStrategyReply, error) {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	query := tx.Strategy
	wrappers := query.WithContext(ctx).Where(query.TeamID.Eq(teamId))
	if validate.TextIsNotNull(params.Keyword) {
		wrappers = wrappers.Where(query.Name.Like(params.Keyword))
	}
	if !params.Status.IsUnknown() {
		wrappers = wrappers.Where(query.Status.Eq(params.Status.GetValue()))
	}
	if len(params.GroupIds) > 0 {
		wrappers = wrappers.Where(query.StrategyGroupID.In(params.GroupIds...))
	}
	if len(params.StrategyTypes) > 0 {
		wrappers = wrappers.Where(query.StrategyType.In(slices.Map(params.StrategyTypes, func(strategyType vobj.StrategyType) int8 { return strategyType.GetValue() })...))
	}
	if validate.IsNotNil(params.PaginationRequest) {
		total, err := wrappers.Count()
		if err != nil {
			return nil, err
		}
		params.WithTotal(total)
		wrappers = wrappers.Limit(int(params.Limit)).Offset(params.Offset())
	}
	strategies, err := wrappers.Preload(query.StrategyGroup).Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(strategies, func(strategy *team.Strategy) do.Strategy { return strategy })
	return params.ToListReply(rows), nil
}

// UpdateStatus implements repository.TeamStrategy.
func (t *teamStrategyRepoImpl) UpdateStatus(ctx context.Context, params *bo.UpdateTeamStrategiesStatusParams) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.Strategy
	wrappers := []gen.Condition{
		mutation.ID.In(params.StrategyIds...),
		mutation.TeamID.Eq(teamId),
	}
	mutations := []field.AssignExpr{
		mutation.Status.Value(params.Status.GetValue()),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).UpdateSimple(mutations...)
	return err
}

func (t *teamStrategyRepoImpl) FindByIds(ctx context.Context, strategyIds []uint32) ([]do.Strategy, error) {
	if len(strategyIds) == 0 {
		return nil, nil
	}
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := bizQuery.Strategy
	wrapper := mutation.WithContext(ctx).Where(mutation.TeamID.Eq(teamId), mutation.ID.In(strategyIds...))
	rows, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(rows, func(row *team.Strategy) do.Strategy { return row }), nil
}

func (t *teamStrategyRepoImpl) Get(ctx context.Context, strategyId uint32) (do.Strategy, error) {
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := bizQuery.Strategy.WithContext(ctx).Where(bizQuery.Strategy.TeamID.Eq(teamId), bizQuery.Strategy.ID.Eq(strategyId))
	strategy, err := wrapper.Preload(field.Associations).First()
	if err != nil {
		return nil, strategyNotFound(err)
	}
	return strategy, nil
}
