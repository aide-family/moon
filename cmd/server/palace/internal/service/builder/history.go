package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	historyapi "github.com/aide-family/moon/api/admin/history"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
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
		WithGetAlarmHistoryRequest(request *historyapi.GetHistoryRequest) IGetAlarmHistoryRequestBuilder
		WithListAlarmHistoryRequest(request *historyapi.ListHistoryRequest) IListAlarmHistoryRequestBuilder
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

	IDoAlarmHistoryBuilder interface {
		ToApi(*alarmmodel.AlarmHistory) *admin.AlarmHistoryItem
		ToApis([]*alarmmodel.AlarmHistory) []*admin.AlarmHistoryItem
	}

	doAlarmHistoryBuilder struct {
		ctx context.Context
	}
)

func (a *doAlarmHistoryBuilder) ToApi(history *alarmmodel.AlarmHistory) *admin.AlarmHistoryItem {
	if types.IsNil(a) || types.IsNil(history) {
		return nil
	}
	return &admin.AlarmHistoryItem{
		Id:           history.ID,
		InstanceName: history.InstanceName,
		AlertStatus:  api.AlertStatus(history.AlertStatus),
		Expr:         history.Expr,
		Fingerprint:  history.Fingerprint,
	}
}

func (a *doAlarmHistoryBuilder) ToApis(histories []*alarmmodel.AlarmHistory) []*admin.AlarmHistoryItem {
	if types.IsNil(a) || types.IsNil(histories) {
		return nil
	}
	return types.SliceTo(histories, func(history *alarmmodel.AlarmHistory) *admin.AlarmHistoryItem {
		return a.ToApi(history)
	})
}

func (l *listAlarmHistoryRequestBuilder) ToBo() *bo.QueryAlarmHistoryListParams {
	if types.IsNil(l) || types.IsNil(l.ListHistoryRequest) {
		return nil
	}
	return &bo.QueryAlarmHistoryListParams{
		Keyword:      l.GetKeyword(),
		AlertStatus:  vobj.AlertStatus(l.AlarmStatuses),
		Page:         types.NewPagination(l.GetPagination()),
		InstanceName: l.GetInstanceName(),
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
