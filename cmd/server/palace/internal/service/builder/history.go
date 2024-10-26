package builder

import (
	"context"
	"time"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	historyapi "github.com/aide-family/moon/api/admin/history"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ IAlarmHistoryModuleBuilder = (*alarmHistoryModuleBuilder)(nil)

type (
	alarmHistoryModuleBuilder struct {
		ctx context.Context
	}

	// IAlarmHistoryModuleBuilder alarm history module builder
	IAlarmHistoryModuleBuilder interface {
		WithGetAlarmHistoryRequest(*historyapi.GetHistoryRequest) IGetAlarmHistoryRequestBuilder
		WithListAlarmHistoryRequest(*historyapi.ListHistoryRequest) IListAlarmHistoryRequestBuilder
		DoAlarmHistoryItemBuilder() IDoAlarmHistoryBuilder
	}

	// IGetAlarmHistoryRequestBuilder get alarm history request builder
	IGetAlarmHistoryRequestBuilder interface {
		ToBo() *bo.GetAlarmHistoryParams
	}

	getAlarmHistoryRequestBuilder struct {
		ctx context.Context
		*historyapi.GetHistoryRequest
	}

	// IListAlarmHistoryRequestBuilder list alarm history request builder
	IListAlarmHistoryRequestBuilder interface {
		ToBo() *bo.QueryAlarmHistoryListParams
	}
	listAlarmHistoryRequestBuilder struct {
		ctx context.Context
		*historyapi.ListHistoryRequest
	}

	// IDoAlarmHistoryBuilder do alarm history builder
	IDoAlarmHistoryBuilder interface {
		ToAPI(*alarmmodel.AlarmHistory) *admin.AlarmHistoryItem
		ToAPIs([]*alarmmodel.AlarmHistory) []*admin.AlarmHistoryItem
	}

	doAlarmHistoryBuilder struct {
		ctx context.Context
	}
)

func (a *doAlarmHistoryBuilder) ToAPI(history *alarmmodel.AlarmHistory) *admin.AlarmHistoryItem {
	if types.IsNil(a) || types.IsNil(history) {
		return nil
	}

	resItem := &admin.AlarmHistoryItem{
		Id:          history.ID,
		StartsAt:    history.StartsAt,
		EndsAt:      history.EndsAt,
		AlertStatus: api.AlertStatus(history.AlertStatus),
		Level:       nil,
		Strategy:    nil,
		Description: history.Description,
		Expr:        history.Expr,
		Datasource:  nil,
		Fingerprint: history.Fingerprint,
		RawInfo:     "",
		Labels:      history.Labels.Map(),
		Annotations: history.Annotations,
		Summary:     history.Summary,
		Duration: types.Ternary(
			history.AlertStatus.IsResolved(),
			types.NewTimeByString(history.EndsAt).Sub(types.NewTimeByString(history.StartsAt).Time).String(),
			time.Since(types.NewTimeByString(history.StartsAt).Time).String(),
		),
	}

	details := history.HistoryDetails
	if !types.IsNil(details) {
		datasource := &bizmodel.Datasource{}
		_ = datasource.UnmarshalBinary([]byte(details.Datasource))
		resItem.Datasource = NewParamsBuild().DatasourceModuleBuilder().DoDatasourceBuilder().ToAPI(datasource)

		strategy := &bizmodel.Strategy{}
		_ = strategy.UnmarshalBinary([]byte(details.Strategy))
		resItem.Strategy = NewParamsBuild().StrategyModuleBuilder().DoStrategyBuilder().ToAPI(strategy)

		level := &bizmodel.StrategyLevel{}
		_ = level.UnmarshalBinary([]byte(details.Level))
		resItem.Level = NewParamsBuild().StrategyModuleBuilder().DoStrategyLevelBuilder().ToAPI(level)
	}

	return resItem
}

func (a *doAlarmHistoryBuilder) ToAPIs(histories []*alarmmodel.AlarmHistory) []*admin.AlarmHistoryItem {
	if types.IsNil(a) || types.IsNil(histories) {
		return nil
	}
	return types.SliceTo(histories, func(history *alarmmodel.AlarmHistory) *admin.AlarmHistoryItem {
		return a.ToAPI(history)
	})
}

func (l *listAlarmHistoryRequestBuilder) ToBo() *bo.QueryAlarmHistoryListParams {
	if types.IsNil(l) || types.IsNil(l.ListHistoryRequest) {
		return nil
	}
	return &bo.QueryAlarmHistoryListParams{
		Keyword: l.GetKeyword(),
		AlertStatus: types.SliceTo(l.AlarmStatuses, func(status api.AlertStatus) vobj.AlertStatus {
			return vobj.AlertStatus(status)
		}),
		Page:            types.NewPagination(l.GetPagination()),
		EventAtStart:    l.GetEventAtStart(),
		EventAtEnd:      l.GetEventAtEnd(),
		ResolvedAtStart: l.GetRecoverAtStart(),
		ResolvedAtEnd:   l.GetRecoverAtEnd(),
		AlarmPage:       l.GetAlarmPage(),
	}
}

func (a *alarmHistoryModuleBuilder) WithListAlarmHistoryRequest(request *historyapi.ListHistoryRequest) IListAlarmHistoryRequestBuilder {
	return &listAlarmHistoryRequestBuilder{
		ctx:                a.ctx,
		ListHistoryRequest: request,
	}
}

func (a *getAlarmHistoryRequestBuilder) ToBo() *bo.GetAlarmHistoryParams {
	if types.IsNil(a) || types.IsNil(a.GetHistoryRequest) {
		return nil
	}
	return &bo.GetAlarmHistoryParams{
		ID: a.GetId(),
	}
}

func (a *alarmHistoryModuleBuilder) WithGetAlarmHistoryRequest(request *historyapi.GetHistoryRequest) IGetAlarmHistoryRequestBuilder {
	return &getAlarmHistoryRequestBuilder{
		ctx:               a.ctx,
		GetHistoryRequest: request,
	}
}
func (a *alarmHistoryModuleBuilder) DoAlarmHistoryBuilder() IDoAlarmHistoryBuilder {
	return &doAlarmHistoryBuilder{
		ctx: a.ctx,
	}
}

func (a *alarmHistoryModuleBuilder) DoAlarmHistoryItemBuilder() IDoAlarmHistoryBuilder {
	return &doAlarmHistoryBuilder{
		ctx: a.ctx,
	}
}
