package build

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	alarmapi "github.com/aide-family/moon/api/admin/alarm"
	realtimeapi "github.com/aide-family/moon/api/admin/realtime"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtimecache"
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

	// AlarmGroupModuleBuilder 告警组模块构造器
	AlarmGroupModuleBuilder interface {
		WithDoAlarmGroup(*bizmodel.AlarmNoticeGroup) DoAlarmGroupBuilder
		WithDosAlarmGroup([]*bizmodel.AlarmNoticeGroup) DosAlarmGroupBuilder
		WithAPICreateAlarmGroupRequest(item *alarmapi.CreateAlarmGroupRequest) APICreateAlarmGroupParamsBuilder
		WithAPIQueryAlarmGroupListRequest(request *alarmapi.ListAlarmGroupRequest) APIQueryAlarmGroupListParamsBuilder
		WithAPIUpdateAlarmGroupRequest(*alarmapi.UpdateAlarmGroupRequest) APIUpdateAlarmGroupParamsBuilder
		WithAPIUpdateStatusAlarmGroupRequest(request *alarmapi.UpdateAlarmGroupStatusRequest) APIUpdateAlarmGroupStatusParamsBuilder
		WithContext(context.Context) AlarmGroupModuleBuilder
	}

	// DoAlarmGroupBuilder do  alarm group builder
	DoAlarmGroupBuilder interface {
		ToAPI() *adminapi.AlarmNoticeGroupItem
		ToSelect() *adminapi.SelectItem
	}

	// DosAlarmGroupBuilder do  alarm group builder
	DosAlarmGroupBuilder interface {
		ToAPIs() []*adminapi.AlarmNoticeGroupItem
		ToSelects() []*adminapi.SelectItem
	}

	// APICreateAlarmGroupParamsBuilder create alarm group params builder
	APICreateAlarmGroupParamsBuilder interface {
		ToBo() *bo.CreateAlarmGroupParams
	}

	// APIQueryAlarmGroupListParamsBuilder query alarm group list params builder
	APIQueryAlarmGroupListParamsBuilder interface {
		ToBo() *bo.QueryAlarmGroupListParams
	}

	// APIUpdateAlarmGroupParamsBuilder update alarm group params builder
	APIUpdateAlarmGroupParamsBuilder interface {
		ToBo() *bo.UpdateAlarmGroupParams
	}

	// APIUpdateAlarmGroupStatusParamsBuilder update alarm group status params builder
	APIUpdateAlarmGroupStatusParamsBuilder interface {
		ToBo() *bo.UpdateAlarmGroupStatusParams
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

	// alarm group
	doAlarmGroupBuilder struct {
		alarm *bizmodel.AlarmNoticeGroup
		ctx   context.Context
	}

	dosAlarmGroupBuilder struct {
		alarms []*bizmodel.AlarmNoticeGroup
		ctx    context.Context
	}

	alarmModuleBuilder struct {
		ctx context.Context
	}

	apiCreateAlarmGroupParamsBuilder struct {
		ctx   context.Context
		param *alarmapi.CreateAlarmGroupRequest
	}

	apiUpdateAlarmGroupParamsBuilder struct {
		ctx   context.Context
		param *alarmapi.UpdateAlarmGroupRequest
	}

	apiUpdateAlarmGroupStatusParamsBuilder struct {
		ctx   context.Context
		param *alarmapi.UpdateAlarmGroupStatusRequest
	}

	apiListAlarmGroupParamsBuilder struct {
		ctx   context.Context
		param *alarmapi.ListAlarmGroupRequest
	}
)

func (a *doAlarmGroupBuilder) ToSelect() *adminapi.SelectItem {
	if types.IsNil(a) {
		return nil
	}
	return &adminapi.SelectItem{
		Value:    a.alarm.ID,
		Label:    a.alarm.Name,
		Children: nil,
		Disabled: a.alarm.DeletedAt > 0 || !a.alarm.Status.IsEnable(),
		Extend: &adminapi.SelectExtend{
			Remark: a.alarm.Remark,
		},
	}
}

func (a *dosAlarmGroupBuilder) ToSelects() []*adminapi.SelectItem {
	if types.IsNil(a) || types.IsNil(a.alarms) {
		return nil
	}
	return types.SliceTo(a.alarms, func(item *bizmodel.AlarmNoticeGroup) *adminapi.SelectItem {
		groupSelectInfo := NewBuilder().AlarmGroupModule().
			WithContext(a.ctx).WithDoAlarmGroup(item).ToSelect()
		return groupSelectInfo
	})
}

func (a *dosAlarmGroupBuilder) ToAPIs() []*adminapi.AlarmNoticeGroupItem {
	if types.IsNil(a) || types.IsNil(a.alarms) {
		return nil
	}
	return types.SliceTo(a.alarms, func(item *bizmodel.AlarmNoticeGroup) *adminapi.AlarmNoticeGroupItem {
		alarmGroup := NewBuilder().AlarmGroupModule().
			WithContext(a.ctx).WithDoAlarmGroup(item).ToAPI()
		return alarmGroup
	})
}

func (a *alarmModuleBuilder) WithDoAlarmGroup(group *bizmodel.AlarmNoticeGroup) DoAlarmGroupBuilder {
	return newDoAlarmGroupBuilder(a.ctx, group)
}

func (a *alarmModuleBuilder) WithDosAlarmGroup(groups []*bizmodel.AlarmNoticeGroup) DosAlarmGroupBuilder {
	return newDosAlarmGroupBuilder(a.ctx, groups)
}

func (a *alarmModuleBuilder) WithAPICreateAlarmGroupRequest(request *alarmapi.CreateAlarmGroupRequest) APICreateAlarmGroupParamsBuilder {
	return newAPICreateAlarmParamsBuilder(a.ctx, request)
}

func (a *alarmModuleBuilder) WithAPIQueryAlarmGroupListRequest(request *alarmapi.ListAlarmGroupRequest) APIQueryAlarmGroupListParamsBuilder {
	return newAPIListAlarmGroupParamsBuilder(a.ctx, request)
}

func (a *alarmModuleBuilder) WithAPIUpdateAlarmGroupRequest(request *alarmapi.UpdateAlarmGroupRequest) APIUpdateAlarmGroupParamsBuilder {
	return newAPIUpdateAlarmGroupParamsBuilder(a.ctx, request)
}

func (a *alarmModuleBuilder) WithAPIUpdateStatusAlarmGroupRequest(request *alarmapi.UpdateAlarmGroupStatusRequest) APIUpdateAlarmGroupStatusParamsBuilder {
	return newAPIUpdateAlarmGroupStatusParamsBuilder(a.ctx, request)
}

func (a *alarmModuleBuilder) WithContext(ctx context.Context) AlarmGroupModuleBuilder {
	if types.IsNil(a) {
		return newAlarmModuleBuilder(ctx)
	}
	a.ctx = ctx
	return a
}

func (a *doAlarmGroupBuilder) ToAPI() *adminapi.AlarmNoticeGroupItem {
	if types.IsNil(a) || types.IsNil(a.alarm) {
		return nil
	}
	cache := runtimecache.GetRuntimeCache()
	alarmGroup := &adminapi.AlarmNoticeGroupItem{
		Id:        a.alarm.ID,
		Name:      a.alarm.Name,
		Remark:    a.alarm.Remark,
		Status:    api.Status(a.alarm.Status),
		CreatedAt: a.alarm.CreatedAt.String(),
		UpdatedAt: a.alarm.UpdatedAt.String(),
		Creator:   NewBuilder().WithAPIUserBo(cache.GetUser(a.ctx, a.alarm.CreatorID)).GetUsername(),
		NoticeUsers: types.SliceTo(a.alarm.NoticeUsers, func(user *bizmodel.AlarmNoticeUser) *adminapi.NoticeItem {
			return &adminapi.NoticeItem{
				User:       NewBuilder().WithAPIUserBo(cache.GetUser(a.ctx, user.UserID)).ToAPI(),
				NotifyType: api.NotifyType(user.AlarmNoticeType),
			}
		}),
		Hooks: types.SliceTo(a.alarm.AlarmHooks, func(hook *bizmodel.AlarmHook) *adminapi.AlarmHookItem {
			return NewBuilder().HookModuleBuilder().WithDoAlarmHook(hook).ToAPI()
		}),
	}
	return alarmGroup
}

func (a *apiListAlarmGroupParamsBuilder) ToBo() *bo.QueryAlarmGroupListParams {
	if types.IsNil(a) || types.IsNil(a.param) {
		return nil
	}
	return &bo.QueryAlarmGroupListParams{
		Keyword: a.param.GetKeyword(),
		Status:  vobj.Status(a.param.GetStatus()),
		Page:    types.NewPagination(a.param.GetPagination()),
	}
}

func (a *apiUpdateAlarmGroupStatusParamsBuilder) ToBo() *bo.UpdateAlarmGroupStatusParams {
	if types.IsNil(a) || types.IsNil(a.param) {
		return nil
	}
	return &bo.UpdateAlarmGroupStatusParams{
		IDs:    a.param.GetIds(),
		Status: vobj.Status(a.param.GetStatus()),
	}
}

func (a *apiUpdateAlarmGroupParamsBuilder) ToBo() *bo.UpdateAlarmGroupParams {
	if types.IsNil(a) || types.IsNil(a.param) {
		return nil
	}
	return &bo.UpdateAlarmGroupParams{
		ID: a.param.GetId(),
		UpdateParam: &bo.CreateAlarmGroupParams{
			Name:   a.param.GetUpdate().GetName(),
			Remark: a.param.GetUpdate().GetRemark(),
			Status: vobj.Status(a.param.GetUpdate().GetStatus()),
			NoticeUsers: types.SliceTo(a.param.GetUpdate().GetNoticeUser(), func(user *alarmapi.CreateNoticeUserRequest) *bo.CreateNoticeUserParams {
				return &bo.CreateNoticeUserParams{
					UserID:     user.GetUserId(),
					NotifyType: vobj.NotifyType(user.GetNotifyType()),
				}
			}),
		},
	}
}

func (a apiCreateAlarmGroupParamsBuilder) ToBo() *bo.CreateAlarmGroupParams {
	if types.IsNil(a) || types.IsNil(a.param) {
		return nil
	}
	return &bo.CreateAlarmGroupParams{
		Name:   a.param.GetName(),
		Remark: a.param.GetRemark(),
		Status: vobj.Status(a.param.GetStatus()),
		NoticeUsers: types.SliceTo(a.param.NoticeUser, func(user *alarmapi.CreateNoticeUserRequest) *bo.CreateNoticeUserParams {
			return &bo.CreateNoticeUserParams{
				UserID:     user.GetUserId(),
				NotifyType: vobj.NotifyType(user.GetNotifyType()),
			}
		}),
		HookIds: a.param.GetHookIds(),
	}
}

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
		AlarmPageID:     params.GetAlarmPage(),
		MyAlarm:         params.GetMyAlarm(),
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

func newAPICreateAlarmParamsBuilder(ctx context.Context, req *alarmapi.CreateAlarmGroupRequest) APICreateAlarmGroupParamsBuilder {
	return &apiCreateAlarmGroupParamsBuilder{ctx: ctx, param: req}
}

func newAPIUpdateAlarmGroupParamsBuilder(ctx context.Context, req *alarmapi.UpdateAlarmGroupRequest) APIUpdateAlarmGroupParamsBuilder {
	return &apiUpdateAlarmGroupParamsBuilder{ctx: ctx, param: req}
}

func newAPIUpdateAlarmGroupStatusParamsBuilder(ctx context.Context, req *alarmapi.UpdateAlarmGroupStatusRequest) APIUpdateAlarmGroupStatusParamsBuilder {
	return &apiUpdateAlarmGroupStatusParamsBuilder{ctx: ctx, param: req}
}

func newAPIListAlarmGroupParamsBuilder(ctx context.Context, req *alarmapi.ListAlarmGroupRequest) APIQueryAlarmGroupListParamsBuilder {
	return &apiListAlarmGroupParamsBuilder{ctx: ctx, param: req}
}

func newDoAlarmGroupBuilder(ctx context.Context, alarmGroup *bizmodel.AlarmNoticeGroup) DoAlarmGroupBuilder {
	return &doAlarmGroupBuilder{ctx: ctx, alarm: alarmGroup}
}

func newDosAlarmGroupBuilder(ctx context.Context, alarmGroups []*bizmodel.AlarmNoticeGroup) DosAlarmGroupBuilder {
	return &dosAlarmGroupBuilder{ctx: ctx, alarms: alarmGroups}
}

func newAlarmModuleBuilder(ctx context.Context) AlarmGroupModuleBuilder {
	return &alarmModuleBuilder{ctx: ctx}
}
