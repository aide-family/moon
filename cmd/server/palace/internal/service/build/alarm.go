package build

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	realtimeapi "github.com/aide-family/moon/api/admin/realtime"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// RealtimeAlarmModuleBuilder 实时告警模块构造器
	RealtimeAlarmModuleBuilder interface {
		WithDoRealtimeAlarm(*bizmodel.RealtimeAlarm) DoRealtimeAlarmBuilder
		WithDostRealtimeAlarm([]*bizmodel.RealtimeAlarm) DostRealtimeAlarmBuilder
		WithContext(context.Context) RealtimeAlarmModuleBuilder
		WithAPIGetAlarmRequest(*realtimeapi.GetAlarmRequest) APIGetRealTimeAlarmParamsBuilder
		WithAPIListAlarmRequest(*realtimeapi.ListAlarmRequest) APIListAlarmParamsBuilder
	}

	// DoRealtimeAlarmBuilder do realtime alarm builder
	DoRealtimeAlarmBuilder interface {
		ToAPI() *adminapi.RealtimeAlarmItem
	}

	// DostRealtimeAlarmBuilder do realtime alarm builder
	DostRealtimeAlarmBuilder interface {
		ToAPIs() []*adminapi.RealtimeAlarmItem
	}

	// APIGetRealTimeAlarmParamsBuilder get realtime alarm params builder
	APIGetRealTimeAlarmParamsBuilder interface {
		ToBo() *bo.GetRealTimeAlarmParams
		WithFingerprint(fingerprint string) APIGetRealTimeAlarmParamsBuilder
	}

	// APIListAlarmParamsBuilder get realtime alarm params builder
	APIListAlarmParamsBuilder interface {
		ToBo() *bo.GetRealTimeAlarmsParams
	}

	doRealtimeAlarmBuilder struct {
		alarm *bizmodel.RealtimeAlarm
		ctx   context.Context
	}

	dostRealtimeAlarmBuilder struct {
		alarms []*bizmodel.RealtimeAlarm
		ctx    context.Context
	}

	realtimeAlarmModuleBuilder struct {
		ctx context.Context
	}

	apiGetRealTimeAlarmParamsBuilder struct {
		ctx         context.Context
		req         *realtimeapi.GetAlarmRequest
		fingerprint string
	}

	apiGetRealTimeAlarmsParamsBuilder struct {
		ctx    context.Context
		params *realtimeapi.ListAlarmRequest
	}
)

func (a *apiGetRealTimeAlarmsParamsBuilder) ToBo() *bo.GetRealTimeAlarmsParams {
	if types.IsNil(a) || types.IsNil(a.params) {
		return nil
	}
	params := a.params
	return &bo.GetRealTimeAlarmsParams{
		Pagination:      types.NewPagination(params.GetPagination()),
		EventAtStart:    params.GetEventAtStart(),
		EventAtEnd:      params.GetEventAtEnd(),
		ResolvedAtStart: params.GetRecoverAtStart(),
		ResolvedAtEnd:   params.GetRecoverAtEnd(),
		AlarmLevels:     params.GetAlarmLevels(),
		AlarmStatuses:   types.SliceTo(params.GetAlarmStatuses(), func(item api.AlertStatus) vobj.AlertStatus { return vobj.AlertStatus(item) }),
		Keyword:         params.GetKeyword(),
	}
}

func (a *apiGetRealTimeAlarmParamsBuilder) WithFingerprint(fingerprint string) APIGetRealTimeAlarmParamsBuilder {
	if types.IsNil(a) || types.IsNil(a.req) {
		return newAPIGetRealTimeAlarmParamsBuilder(context.TODO(), &realtimeapi.GetAlarmRequest{}).WithFingerprint(fingerprint)
	}
	a.fingerprint = fingerprint
	return a
}

func (a *apiGetRealTimeAlarmParamsBuilder) ToBo() *bo.GetRealTimeAlarmParams {
	if types.IsNil(a) || types.IsNil(a.req) {
		return nil
	}
	params := a.req
	return &bo.GetRealTimeAlarmParams{
		RealtimeAlarmID: params.GetId(),
		Fingerprint:     "",
	}
}

func (r *realtimeAlarmModuleBuilder) WithAPIListAlarmRequest(request *realtimeapi.ListAlarmRequest) APIListAlarmParamsBuilder {
	return newAPIListAlarmParamsBuilder(r.ctx, request)
}

func (r *realtimeAlarmModuleBuilder) WithAPIGetAlarmRequest(request *realtimeapi.GetAlarmRequest) APIGetRealTimeAlarmParamsBuilder {
	return newAPIGetRealTimeAlarmParamsBuilder(r.ctx, request)
}

func (r *realtimeAlarmModuleBuilder) WithContext(ctx context.Context) RealtimeAlarmModuleBuilder {
	if types.IsNil(r) {
		return newRealtimeAlarmModuleBuilder(ctx)
	}
	r.ctx = ctx
	return r
}

func (r *realtimeAlarmModuleBuilder) WithDoRealtimeAlarm(alarm *bizmodel.RealtimeAlarm) DoRealtimeAlarmBuilder {
	return newDoRealtimeAlarmBuilder(r.ctx, alarm)
}

func (r *realtimeAlarmModuleBuilder) WithDostRealtimeAlarm(alarms []*bizmodel.RealtimeAlarm) DostRealtimeAlarmBuilder {
	return newDostRealtimeAlarmBuilder(r.ctx, alarms)
}

func (d *dostRealtimeAlarmBuilder) ToAPIs() []*adminapi.RealtimeAlarmItem {
	if types.IsNil(d) || types.IsNil(d.alarms) {
		return nil
	}
	alarms := d.alarms
	return types.SliceTo(alarms, func(alarm *bizmodel.RealtimeAlarm) *adminapi.RealtimeAlarmItem {
		return newDoRealtimeAlarmBuilder(d.ctx, alarm).ToAPI()
	})
}

func (r *doRealtimeAlarmBuilder) ToAPI() *adminapi.RealtimeAlarmItem {
	if types.IsNil(r) || types.IsNil(r.alarm) {
		return nil
	}
	detail := r.alarm
	return &adminapi.RealtimeAlarmItem{
		Id:       detail.ID,
		StartsAt: types.NewTimeByUnix(detail.StartsAt).String(),
		EndsAt:   types.NewTimeByUnix(detail.EndsAt).String(),
		Status:   api.AlertStatus(detail.Status),
		//Level:        NewBuilder().WithDict(detail.Level).ToAPISelect(),
		LevelID:    detail.LevelID,
		StrategyID: detail.StrategyID,
		//Strategy:     NewBuilder().WithContext(r.ctx).WithAPIStrategy(detail.Strategy).ToAPI(),
		Summary:      detail.Summary,
		Description:  detail.Description,
		Expr:         detail.Expr,
		DatasourceID: detail.DatasourceID,
		//Datasource:   NewBuilder().WithDoDatasource(detail.Datasource).ToAPI(),
		Fingerprint: detail.Fingerprint,
	}
}

func newRealtimeAlarmModuleBuilder(ctx context.Context) RealtimeAlarmModuleBuilder {
	return &realtimeAlarmModuleBuilder{ctx: ctx}
}

func newDoRealtimeAlarmBuilder(ctx context.Context, alarm *bizmodel.RealtimeAlarm) DoRealtimeAlarmBuilder {
	return &doRealtimeAlarmBuilder{alarm: alarm, ctx: ctx}
}

func newDostRealtimeAlarmBuilder(ctx context.Context, alarms []*bizmodel.RealtimeAlarm) DostRealtimeAlarmBuilder {
	return &dostRealtimeAlarmBuilder{alarms: alarms, ctx: ctx}
}

func newAPIGetRealTimeAlarmParamsBuilder(ctx context.Context, req *realtimeapi.GetAlarmRequest) APIGetRealTimeAlarmParamsBuilder {
	return &apiGetRealTimeAlarmParamsBuilder{ctx: ctx, req: req}
}

func newAPIListAlarmParamsBuilder(ctx context.Context, params *realtimeapi.ListAlarmRequest) APIListAlarmParamsBuilder {
	return &apiGetRealTimeAlarmsParamsBuilder{ctx: ctx, params: params}
}
