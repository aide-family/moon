package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/types"

	"gorm.io/gen"
	"gorm.io/gen/field"
)

type (
	subscriberStrategyRepository struct {
		data *data.Data
	}
)

// NewSubscriberStrategyRepository 创建策略仓库
func NewSubscriberStrategyRepository(data *data.Data) repository.SubscriberStrategy {
	return &subscriberStrategyRepository{
		data: data,
	}
}

func (s *subscriberStrategyRepository) UserSubscriberStrategy(ctx context.Context, params *bo.SubscriberStrategyParams) error {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	subscriberModel := createSubscriberStrategyToModel(ctx, params)
	return bizQuery.StrategySubscribers.Create(subscriberModel)
}
func (s *subscriberStrategyRepository) UserUnSubscriberStrategy(ctx context.Context, params *bo.UnSubscriberStrategyParams) error {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	_, err = bizQuery.StrategySubscribers.Where(bizQuery.StrategySubscribers.StrategyID.Eq(params.StrategyID), bizQuery.StrategySubscribers.UserID.Eq(params.UserID)).Delete()
	if !types.IsNil(err) {
		return err
	}
	return nil
}

func (s *subscriberStrategyRepository) UserSubscriberStrategyList(ctx context.Context, params *bo.QueryUserSubscriberParams) ([]*bizmodel.StrategySubscribers, error) {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return nil, err
	}
	bizWrapper := bizQuery.StrategySubscribers.WithContext(ctx)

	var wheres []gen.Condition
	if params.NotifyType != 0 {
		wheres = append(wheres, bizQuery.StrategySubscribers.AlarmNoticeType.Eq(params.NotifyType.GetValue()))
	}
	wheres = append(wheres, bizQuery.StrategySubscribers.UserID.Eq(params.UserID))

	bizWrapper = bizWrapper.Where(wheres...).Preload(field.Associations)

	if err = types.WithPageQuery[bizquery.IStrategySubscribersDo](bizWrapper, params.Page); err != nil {
		return nil, err
	}
	return bizWrapper.Order(bizQuery.StrategySubscribers.ID.Desc()).Find()
}

func (s *subscriberStrategyRepository) StrategySubscriberList(ctx context.Context, params *bo.QueryStrategySubscriberParams) ([]*bizmodel.StrategySubscribers, error) {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return nil, err
	}
	bizWrapper := bizQuery.StrategySubscribers.WithContext(ctx)

	var wheres []gen.Condition
	if params.NotifyType != 0 {
		wheres = append(wheres, bizQuery.StrategySubscribers.AlarmNoticeType.Eq(params.NotifyType.GetValue()))
	}
	wheres = append(wheres, bizQuery.StrategySubscribers.StrategyID.Eq(params.StrategyID))

	bizWrapper = bizWrapper.Where(wheres...).Preload(field.Associations)
	if err = types.WithPageQuery[bizquery.IStrategySubscribersDo](bizWrapper, params.Page); err != nil {
		return nil, err
	}
	return bizWrapper.Order(bizQuery.StrategySubscribers.ID.Desc()).Find()
}

func createSubscriberStrategyToModel(ctx context.Context, params *bo.SubscriberStrategyParams) *bizmodel.StrategySubscribers {
	subscriberModel := &bizmodel.StrategySubscribers{
		StrategyID:      params.StrategyID,
		UserID:          params.UserID,
		AlarmNoticeType: params.NotifyType,
	}
	subscriberModel.WithContext(ctx)
	return subscriberModel
}
