package build

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	subscriberapi "github.com/aide-family/moon/api/admin/subscriber"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtimecache"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// SubscriberStrategyModuleBuilder build subscriber strategy module
	SubscriberStrategyModuleBuilder interface {
		WithAPISubscriberStrategyRequest(req *subscriberapi.SubscriberStrategyRequest) APISubscriberStrategyParamsBuilder

		WithAPIUnSubscriberStrategyRequest(req *subscriberapi.UnSubscriberRequest) APISubscriberStrategyParamsBuilder

		WithAPISubscriberStrategyListRequest(req *subscriberapi.StrategySubscriberRequest) APISubscriberStrategyParamsBuilder

		WithAPIUserSubscriberStrategyListRequest(req *subscriberapi.UserSubscriberListRequest) APISubscriberStrategyParamsBuilder

		WithDosSubscriberStrategy(subs []*bizmodel.StrategySubscribers) DosSubscriberStrategyBuilder

		WithDosUserSubscriberStrategy(subs []*bizmodel.StrategySubscribers) DosUserSubscriberStrategyBuilder
	}

	subscriberStrategyModuleBuilder struct {
		ctx context.Context
	}

	// APISubscriberStrategyParamsBuilder build api subscriber strategy params
	APISubscriberStrategyParamsBuilder interface {
		ToSubscriberBo() *bo.SubscriberStrategyParams
		ToUnSubscriberBo() *bo.UnSubscriberStrategyParams
		ToStrategySubscriberListBo() *bo.QueryStrategySubscriberParams
		ToUserSubscriberListBo() *bo.QueryUserSubscriberParams
	}

	apiSubscriberStrategyParamsBuilder struct {
		StrategySubscriber        *subscriberapi.StrategySubscriberRequest
		UnSubscriberRequest       *subscriberapi.UnSubscriberRequest
		SubscriberStrategyRequest *subscriberapi.SubscriberStrategyRequest
		UserSubscriberListRequest *subscriberapi.UserSubscriberListRequest

		ctx context.Context
	}

	// DosSubscriberStrategyBuilder build dos subscriber strategy
	DosSubscriberStrategyBuilder interface {
		ToAPIs() []*adminapi.StrategySubscriberItem
	}

	dosSubscriberStrategyBuilder struct {
		StrategySubscriberList []*bizmodel.StrategySubscribers
		ctx                    context.Context
	}

	// DosUserSubscriberStrategyBuilder build dos user subscriber strategy
	DosUserSubscriberStrategyBuilder interface {
		ToAPIs() []*adminapi.StrategyItem
	}

	dosUserSubscriberStrategyBuilder struct {
		UserSubscriberList []*bizmodel.StrategySubscribers
		ctx                context.Context
	}
)

func (a *apiSubscriberStrategyParamsBuilder) ToSubscriberBo() *bo.SubscriberStrategyParams {
	if types.IsNil(a) || types.IsNil(a.SubscriberStrategyRequest) {
		return nil
	}
	return &bo.SubscriberStrategyParams{
		StrategyID: a.SubscriberStrategyRequest.GetStrategyId(),
		NotifyType: vobj.NotifyType(a.SubscriberStrategyRequest.GetNotifyType()),
	}
}

func (a *apiSubscriberStrategyParamsBuilder) ToUnSubscriberBo() *bo.UnSubscriberStrategyParams {
	if types.IsNil(a) || types.IsNil(a.UnSubscriberRequest) {
		return nil
	}
	return &bo.UnSubscriberStrategyParams{
		StrategyID: a.UnSubscriberRequest.GetStrategyId(),
	}
}

func (a *apiSubscriberStrategyParamsBuilder) ToStrategySubscriberListBo() *bo.QueryStrategySubscriberParams {
	if types.IsNil(a) || types.IsNil(a.StrategySubscriber) {
		return nil
	}
	return &bo.QueryStrategySubscriberParams{
		StrategyID: a.StrategySubscriber.GetStrategyId(),
		NotifyType: vobj.NotifyType(a.StrategySubscriber.GetNotifyType()),
		Page:       types.NewPagination(a.StrategySubscriber.GetPagination()),
	}
}

func (a *apiSubscriberStrategyParamsBuilder) ToUserSubscriberListBo() *bo.QueryUserSubscriberParams {
	if types.IsNil(a) || types.IsNil(a.UserSubscriberListRequest) {
		return nil
	}
	return &bo.QueryUserSubscriberParams{
		NotifyType: vobj.NotifyType(a.UserSubscriberListRequest.GetNotifyType()),
		Page:       types.NewPagination(a.UserSubscriberListRequest.GetPagination()),
	}
}

func (a *dosSubscriberStrategyBuilder) ToAPIs() []*adminapi.StrategySubscriberItem {
	if types.IsNil(a) || types.IsNil(a.StrategySubscriberList) {
		return nil
	}
	strategySubscriberList := a.StrategySubscriberList
	cache := runtimecache.GetRuntimeCache()
	return types.SliceTo(strategySubscriberList, func(sub *bizmodel.StrategySubscribers) *adminapi.StrategySubscriberItem {
		return &adminapi.StrategySubscriberItem{
			Id:         sub.ID,
			NotifyType: api.NotifyType(sub.AlarmNoticeType),
			User:       NewBuilder().WithAPIUserBo(cache.GetUser(a.ctx, sub.UserID)).ToAPI(),
		}
	})
}

func (a *dosUserSubscriberStrategyBuilder) ToAPIs() []*adminapi.StrategyItem {
	if types.IsNil(a) || types.IsNil(a.UserSubscriberList) {
		return nil
	}
	userSubscriberList := a.UserSubscriberList
	return types.SliceTo(userSubscriberList, func(sub *bizmodel.StrategySubscribers) *adminapi.StrategyItem {
		return NewBuilder().WithAPIStrategy(sub.Strategy).ToAPI()
	})
}

func (a *subscriberStrategyModuleBuilder) WithAPISubscriberStrategyRequest(req *subscriberapi.SubscriberStrategyRequest) APISubscriberStrategyParamsBuilder {
	return &apiSubscriberStrategyParamsBuilder{
		SubscriberStrategyRequest: req,
		ctx:                       a.ctx,
	}
}

func (a *subscriberStrategyModuleBuilder) WithAPIUnSubscriberStrategyRequest(req *subscriberapi.UnSubscriberRequest) APISubscriberStrategyParamsBuilder {
	return &apiSubscriberStrategyParamsBuilder{
		UnSubscriberRequest: req,
		ctx:                 a.ctx,
	}
}

func (a *subscriberStrategyModuleBuilder) WithAPISubscriberStrategyListRequest(req *subscriberapi.StrategySubscriberRequest) APISubscriberStrategyParamsBuilder {
	return &apiSubscriberStrategyParamsBuilder{
		StrategySubscriber: req,
		ctx:                a.ctx,
	}
}

func (a *subscriberStrategyModuleBuilder) WithAPIUserSubscriberStrategyListRequest(req *subscriberapi.UserSubscriberListRequest) APISubscriberStrategyParamsBuilder {
	return &apiSubscriberStrategyParamsBuilder{
		UserSubscriberListRequest: req,
		ctx:                       a.ctx,
	}
}

func (a *subscriberStrategyModuleBuilder) WithDosUserSubscriberStrategy(subs []*bizmodel.StrategySubscribers) DosUserSubscriberStrategyBuilder {
	return &dosUserSubscriberStrategyBuilder{
		UserSubscriberList: subs,
		ctx:                a.ctx,
	}
}
func (a *subscriberStrategyModuleBuilder) WithDosSubscriberStrategy(subs []*bizmodel.StrategySubscribers) DosSubscriberStrategyBuilder {
	return &dosSubscriberStrategyBuilder{
		StrategySubscriberList: subs,
		ctx:                    a.ctx,
	}
}
func (a *subscriberStrategyModuleBuilder) WithContext(ctx context.Context) SubscriberStrategyModuleBuilder {
	if types.IsNil(a) {
		return newSubscriberStrategyModuleBuilder(ctx)
	}
	a.ctx = ctx
	return a
}

func newSubscriberStrategyModuleBuilder(ctx context.Context) SubscriberStrategyModuleBuilder {
	return &subscriberStrategyModuleBuilder{
		ctx: ctx,
	}
}
