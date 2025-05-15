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

func NewTeamStrategyRepo(data *data.Data) repository.TeamStrategy {
	return &teamStrategyImpl{
		Data: data,
	}
}

type teamStrategyImpl struct {
	*data.Data
}

// Create implements repository.TeamStrategy.
func (t *teamStrategyImpl) Create(ctx context.Context, params bo.CreateTeamStrategyParams) (do.Strategy, error) {
	strategyDo := &team.Strategy{
		StrategyGroupID: params.GetStrategyGroup().GetID(),
		Name:            params.GetName(),
		Remark:          params.GetRemark(),
		Status:          vobj.GlobalStatusDisable,
		StrategyType:    params.GetStrategyType(),
		StrategyGroup:   build.ToStrategyGroup(ctx, params.GetStrategyGroup()),
		Notices:         build.ToStrategyNotices(ctx, params.GetReceiverRoutes()),
	}
	strategyDo.WithContext(ctx)
	tx := getTeamBizQuery(ctx, t)
	if err := tx.Strategy.WithContext(ctx).Create(strategyDo); err != nil {
		return nil, err
	}
	if len(strategyDo.Notices) > 0 {
		notice := tx.Strategy.Notices.WithContext(ctx).Model(strategyDo)
		if err := notice.Append(strategyDo.Notices...); err != nil {
			return nil, err
		}
	}
	return strategyDo, nil
}

// Delete implements repository.TeamStrategy.
func (t *teamStrategyImpl) Delete(ctx context.Context, params *bo.OperateTeamStrategyParams) error {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := tx.Strategy
	wrappers := []gen.Condition{
		mutation.ID.Eq(params.StrategyId),
		mutation.TeamID.Eq(teamId),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).Delete()
	return err
}

// Subscribe implements repository.TeamStrategy.
func (t *teamStrategyImpl) Subscribe(ctx context.Context, params bo.SubscribeTeamStrategy) error {
	tx := getTeamBizQuery(ctx, t)
	subscriberDo := &team.StrategySubscriber{
		StrategyID:    params.GetStrategy().GetID(),
		Strategy:      build.ToStrategy(ctx, params.GetStrategy()),
		SubscribeType: params.GetNoticeType(),
	}
	subscriberDo.WithContext(ctx)

	if err := tx.StrategySubscriber.WithContext(ctx).Create(subscriberDo); err != nil {
		return err
	}
	return nil
}

// SubscribeList implements repository.TeamStrategy.
func (t *teamStrategyImpl) SubscribeList(ctx context.Context, params *bo.SubscribeTeamStrategiesParams) (*bo.SubscribeTeamStrategiesReply, error) {
	tx, teamId := getTeamBizQueryWithTeamID(ctx, t)
	query := tx.StrategySubscriber
	wrappers := query.WithContext(ctx).Where(query.TeamID.Eq(teamId))
	if validate.IsNotNil(params.StrategyId) {
		wrappers = wrappers.Where(query.StrategyID.Eq(params.StrategyId))
	}
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
	return params.ToSubscribeTeamStrategiesReply(subscribers), nil
}

// Update implements repository.TeamStrategy.
func (t *teamStrategyImpl) Update(ctx context.Context, params bo.UpdateTeamStrategyParams) (do.Strategy, error) {
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
		return nil, err
	}
	notices := build.ToStrategyNotices(ctx, params.GetReceiverRoutes())
	notice := tx.Strategy.Notices.WithContext(ctx).Model(&team.Strategy{TeamModel: build.ToTeamModel(ctx, params.GetStrategy())})
	if len(notices) > 0 {
		if err := notice.Replace(notices...); err != nil {
			return nil, err
		}
	} else {
		if err := notice.Clear(); err != nil {
			return nil, err
		}
	}

	strategy, err := tx.Strategy.WithContext(ctx).Where(wrappers...).First()
	if err != nil {
		return nil, strategyNotFound(err)
	}
	return strategy, nil
}

// List implements repository.TeamStrategy.
func (t *teamStrategyImpl) List(ctx context.Context, params *bo.ListTeamStrategyParams) (*bo.ListTeamStrategyReply, error) {
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
	return params.ToListTeamStrategyReply(strategies), nil
}

// UpdateStatus implements repository.TeamStrategy.
func (t *teamStrategyImpl) UpdateStatus(ctx context.Context, params *bo.UpdateTeamStrategiesStatusParams) error {
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

func (t *teamStrategyImpl) FindByIds(ctx context.Context, strategyIds []uint32) ([]do.Strategy, error) {
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

func (t *teamStrategyImpl) Get(ctx context.Context, params *bo.OperateTeamStrategyParams) (do.Strategy, error) {
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := bizQuery.Strategy.WithContext(ctx).Where(bizQuery.Strategy.TeamID.Eq(teamId), bizQuery.Strategy.ID.Eq(params.StrategyId))
	strategy, err := wrapper.First()
	if err != nil {
		return nil, strategyNotFound(err)
	}
	return strategy, nil
}
