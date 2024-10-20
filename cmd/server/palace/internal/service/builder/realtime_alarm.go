package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	realtimeapi "github.com/aide-family/moon/api/admin/realtime"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ IRealtimeAlarmModuleBuilder = (*realtimeAlarmModuleBuilder)(nil)

type (
	realtimeAlarmModuleBuilder struct {
		ctx context.Context
	}

	IRealtimeAlarmModuleBuilder interface {
		WithGetAlarmRequest(*realtimeapi.GetAlarmRequest) IGetAlarmRequestBuilder
		WithListAlarmRequest(*realtimeapi.ListAlarmRequest) IListAlarmRequestBuilder
		DoRealtimeAlarmBuilder() IDoRealtimeAlarmBuilder

		DoAlarmPageSelfBuilder() IDoAlarmPageSelfBuilder

		WithCreateDashboardRequest(*realtimeapi.CreateDashboardRequest) ICreateDashboardRequestBuilder
		WithUpdateDashboardRequest(*realtimeapi.UpdateDashboardRequest) IUpdateDashboardRequestBuilder
		WithDeleteDashboardRequest(*realtimeapi.DeleteDashboardRequest) IDeleteDashboardRequestBuilder
		WithListDashboardRequest(*realtimeapi.ListDashboardRequest) IListDashboardRequestBuilder
		DoDashboardBuilder() IDoDashboardBuilder

		BoChartBuilder() IBoChartBuilder
		DoChartBuilder() IDoChartBuilder

		WithBoAddDashboardParams(*bo.AddDashboardParams) IBoAddDashboardParamsBuilder
		WithBoUpdateDashboardParams(*bo.UpdateDashboardParams) IBoUpdateDashboardParamsBuilder
	}

	IBoAddDashboardParamsBuilder interface {
		ToDo() *bizmodel.Dashboard
		WithDashboardID(uint32) IBoAddDashboardParamsBuilder
		ToDoStrategyGroups() []*bizmodel.StrategyGroup
		ToDoCharts() []*bizmodel.DashboardChart
	}

	boAddDashboardParamsBuilder struct {
		ctx         context.Context
		dashboardID uint32
		*bo.AddDashboardParams
	}

	IBoUpdateDashboardParamsBuilder interface {
		ToDo() *bizmodel.Dashboard
		WithDashboardID(uint32) IBoUpdateDashboardParamsBuilder
		ToDoStrategyGroups() []*bizmodel.StrategyGroup
		ToDoCharts() []*bizmodel.DashboardChart
	}

	boUpdateDashboardParamsBuilder struct {
		ctx         context.Context
		dashboardID uint32
		*bo.UpdateDashboardParams
	}

	IGetAlarmRequestBuilder interface {
		ToBo() *bo.GetRealTimeAlarmParams
	}

	getAlarmRequestBuilder struct {
		ctx context.Context
		*realtimeapi.GetAlarmRequest
	}

	IListAlarmRequestBuilder interface {
		ToBo() *bo.GetRealTimeAlarmsParams
	}

	listAlarmRequestBuilder struct {
		ctx context.Context
		*realtimeapi.ListAlarmRequest
	}

	IDoRealtimeAlarmBuilder interface {
		ToAPI(*alarmmodel.RealtimeAlarm) *adminapi.RealtimeAlarmItem
		ToAPIs([]*alarmmodel.RealtimeAlarm) []*adminapi.RealtimeAlarmItem
	}

	doRealtimeAlarmBuilder struct {
		ctx context.Context
	}

	IDoAlarmPageSelfBuilder interface {
		ToAPI(*bizmodel.AlarmPageSelf) *adminapi.SelfAlarmPageItem
		ToAPIs([]*bizmodel.AlarmPageSelf) []*adminapi.SelfAlarmPageItem
	}

	doAlarmPageSelfBuilder struct {
		ctx context.Context
	}

	ICreateDashboardRequestBuilder interface {
		ToBo() *bo.AddDashboardParams
	}

	createDashboardRequestBuilder struct {
		ctx context.Context
		*realtimeapi.CreateDashboardRequest
	}

	IUpdateDashboardRequestBuilder interface {
		ToBo() *bo.UpdateDashboardParams
	}

	updateDashboardRequestBuilder struct {
		ctx context.Context
		*realtimeapi.UpdateDashboardRequest
	}

	IDeleteDashboardRequestBuilder interface {
		ToBo() *bo.DeleteDashboardParams
	}

	deleteDashboardRequestBuilder struct {
		ctx context.Context
		*realtimeapi.DeleteDashboardRequest
	}

	IListDashboardRequestBuilder interface {
		ToBo() *bo.ListDashboardParams
	}

	listDashboardRequestBuilder struct {
		ctx context.Context
		*realtimeapi.ListDashboardRequest
	}

	IDoDashboardBuilder interface {
		ToAPI(*bizmodel.Dashboard) *adminapi.DashboardItem
		ToAPIs([]*bizmodel.Dashboard) []*adminapi.DashboardItem
		ToSelect(*bizmodel.Dashboard) *adminapi.SelectItem
		ToSelects([]*bizmodel.Dashboard) []*adminapi.SelectItem
	}

	doDashboardBuilder struct {
		ctx context.Context
	}

	IBoChartBuilder interface {
		ToBo(*adminapi.ChartItem) *bo.ChartItem
		ToBos([]*adminapi.ChartItem) []*bo.ChartItem
		WithDashboardID(uint32) IBoChartBuilder
	}

	boChartBuilder struct {
		ctx         context.Context
		dashboardID uint32
	}

	IDoChartBuilder interface {
		ToAPI(*bizmodel.DashboardChart) *adminapi.ChartItem
		ToAPIs([]*bizmodel.DashboardChart) []*adminapi.ChartItem
	}

	doChartBuilder struct {
		ctx context.Context
	}
)

func (b *boUpdateDashboardParamsBuilder) ToDo() *bizmodel.Dashboard {
	if types.IsNil(b) || types.IsNil(b.UpdateDashboardParams) {
		return nil
	}

	return &bizmodel.Dashboard{
		AllFieldModel:  model.AllFieldModel{ID: b.dashboardID},
		Name:           b.Name,
		Status:         vobj.StatusEnable,
		Remark:         b.Remark,
		Color:          b.Color,
		Charts:         b.WithDashboardID(0).ToDoCharts(),
		StrategyGroups: b.WithDashboardID(0).ToDoStrategyGroups(),
	}
}

func (b *boUpdateDashboardParamsBuilder) WithDashboardID(u uint32) IBoUpdateDashboardParamsBuilder {
	if !types.IsNil(b) {
		b.dashboardID = u
	}

	return b
}

func (b *boUpdateDashboardParamsBuilder) ToDoStrategyGroups() []*bizmodel.StrategyGroup {
	if types.IsNil(b) {
		return nil
	}

	return types.SliceTo(b.StrategyGroups, func(strategyGroupId uint32) *bizmodel.StrategyGroup {
		return &bizmodel.StrategyGroup{
			AllFieldModel: model.AllFieldModel{ID: strategyGroupId},
		}
	})
}

func (b *boUpdateDashboardParamsBuilder) ToDoCharts() []*bizmodel.DashboardChart {
	if types.IsNil(b) {
		return nil
	}

	return types.SliceTo(b.Charts, func(chartItem *bo.ChartItem) *bizmodel.DashboardChart {
		return &bizmodel.DashboardChart{
			AllFieldModel: model.AllFieldModel{ID: chartItem.ID},
			Name:          chartItem.Name,
			Status:        chartItem.Status,
			Remark:        chartItem.Remark,
			URL:           chartItem.URL,
			DashboardID:   b.dashboardID,
			ChartType:     chartItem.ChartType,
			Width:         chartItem.Width,
			Height:        chartItem.Height,
		}
	})
}

func (b *boAddDashboardParamsBuilder) ToDo() *bizmodel.Dashboard {
	if types.IsNil(b) || types.IsNil(b.AddDashboardParams) {
		return nil
	}

	return &bizmodel.Dashboard{
		AllFieldModel:  model.AllFieldModel{ID: b.dashboardID},
		Name:           b.Name,
		Status:         vobj.StatusEnable,
		Remark:         b.Remark,
		Color:          b.Color,
		Charts:         b.WithDashboardID(0).ToDoCharts(),
		StrategyGroups: b.WithDashboardID(0).ToDoStrategyGroups(),
	}
}

func (b *boAddDashboardParamsBuilder) WithDashboardID(u uint32) IBoAddDashboardParamsBuilder {
	if !types.IsNil(b) {
		b.dashboardID = u
	}

	return b
}

func (b *boAddDashboardParamsBuilder) ToDoStrategyGroups() []*bizmodel.StrategyGroup {
	if types.IsNil(b) {
		return nil
	}

	return types.SliceTo(b.StrategyGroups, func(strategyGroupId uint32) *bizmodel.StrategyGroup {
		return &bizmodel.StrategyGroup{
			AllFieldModel: model.AllFieldModel{ID: strategyGroupId},
		}
	})
}

func (b *boAddDashboardParamsBuilder) ToDoCharts() []*bizmodel.DashboardChart {
	if types.IsNil(b) {
		return nil
	}

	return types.SliceTo(b.Charts, func(chartItem *bo.ChartItem) *bizmodel.DashboardChart {
		return &bizmodel.DashboardChart{
			AllFieldModel: model.AllFieldModel{ID: chartItem.ID},
			Name:          chartItem.Name,
			Status:        chartItem.Status,
			Remark:        chartItem.Remark,
			URL:           chartItem.URL,
			DashboardID:   b.dashboardID,
			ChartType:     chartItem.ChartType,
			Width:         chartItem.Width,
			Height:        chartItem.Height,
		}
	})
}

func (r *realtimeAlarmModuleBuilder) WithBoAddDashboardParams(params *bo.AddDashboardParams) IBoAddDashboardParamsBuilder {
	if types.IsNil(r) || types.IsNil(params) {
		return nil
	}

	return &boAddDashboardParamsBuilder{ctx: r.ctx, AddDashboardParams: params}
}

func (r *realtimeAlarmModuleBuilder) WithBoUpdateDashboardParams(params *bo.UpdateDashboardParams) IBoUpdateDashboardParamsBuilder {
	if types.IsNil(r) || types.IsNil(params) {
		return nil
	}

	return &boUpdateDashboardParamsBuilder{ctx: r.ctx, UpdateDashboardParams: params, dashboardID: params.ID}
}

func (b *boChartBuilder) WithDashboardID(u uint32) IBoChartBuilder {
	if !types.IsNil(b) {
		b.dashboardID = u
	}
	return b
}

func (b *boChartBuilder) ToBo(item *adminapi.ChartItem) *bo.ChartItem {
	if types.IsNil(b) || types.IsNil(item) {
		return nil
	}

	return &bo.ChartItem{
		ID:          item.GetId(),
		Name:        item.GetTitle(),
		Remark:      item.GetRemark(),
		URL:         item.GetUrl(),
		Status:      vobj.Status(item.GetStatus()),
		Height:      item.GetHeight(),
		Width:       item.GetWidth(),
		ChartType:   vobj.DashboardChartType(item.GetChartType()),
		DashboardID: b.dashboardID,
	}
}

func (b *boChartBuilder) ToBos(items []*adminapi.ChartItem) []*bo.ChartItem {
	if types.IsNil(b) || types.IsNil(items) {
		return nil
	}

	return types.SliceTo(items, func(item *adminapi.ChartItem) *bo.ChartItem {
		return b.ToBo(item)
	})
}

func (r *realtimeAlarmModuleBuilder) BoChartBuilder() IBoChartBuilder {
	return &boChartBuilder{ctx: r.ctx}
}

func (d *doChartBuilder) ToAPI(chart *bizmodel.DashboardChart) *adminapi.ChartItem {
	if types.IsNil(d) || types.IsNil(chart) {
		return nil
	}

	return &adminapi.ChartItem{
		Id:        chart.ID,
		Title:     chart.Name,
		Remark:    chart.Remark,
		Url:       chart.URL,
		Status:    api.Status(chart.Status),
		ChartType: api.ChartType(chart.ChartType),
		Width:     chart.Width,
		Height:    chart.Height,
	}
}

func (d *doChartBuilder) ToAPIs(charts []*bizmodel.DashboardChart) []*adminapi.ChartItem {
	if types.IsNil(d) || types.IsNil(charts) {
		return nil
	}

	return types.SliceTo(charts, func(chart *bizmodel.DashboardChart) *adminapi.ChartItem {
		return d.ToAPI(chart)
	})
}

func (r *realtimeAlarmModuleBuilder) DoChartBuilder() IDoChartBuilder {
	return &doChartBuilder{ctx: r.ctx}
}

func (d *doDashboardBuilder) ToAPI(dashboard *bizmodel.Dashboard) *adminapi.DashboardItem {
	if types.IsNil(d) || types.IsNil(dashboard) {
		return nil
	}

	return &adminapi.DashboardItem{
		Id:        dashboard.ID,
		Title:     dashboard.Name,
		Remark:    dashboard.Remark,
		CreatedAt: dashboard.CreatedAt.String(),
		UpdatedAt: dashboard.UpdatedAt.String(),
		Color:     dashboard.Color,
		Charts:    NewParamsBuild().WithContext(d.ctx).RealtimeAlarmModuleBuilder().DoChartBuilder().ToAPIs(dashboard.Charts),
		Status:    api.Status(dashboard.Status),
		Groups:    NewParamsBuild().WithContext(d.ctx).StrategyModuleBuilder().DoStrategyGroupBuilder().ToAPIs(dashboard.StrategyGroups),
	}
}

func (d *doDashboardBuilder) ToAPIs(dashboards []*bizmodel.Dashboard) []*adminapi.DashboardItem {
	if types.IsNil(d) || types.IsNil(dashboards) {
		return nil
	}

	return types.SliceTo(dashboards, func(dashboard *bizmodel.Dashboard) *adminapi.DashboardItem {
		return d.ToAPI(dashboard)
	})
}

func (d *doDashboardBuilder) ToSelect(dashboard *bizmodel.Dashboard) *adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(dashboard) {
		return nil
	}

	return &adminapi.SelectItem{
		Value:    dashboard.ID,
		Label:    dashboard.Name,
		Children: nil,
		Disabled: dashboard.DeletedAt > 0 || !dashboard.Status.IsEnable(),
		Extend: &adminapi.SelectExtend{
			Color:  dashboard.Color,
			Remark: dashboard.Remark,
		},
	}
}

func (d *doDashboardBuilder) ToSelects(dashboards []*bizmodel.Dashboard) []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(dashboards) {
		return nil
	}

	return types.SliceTo(dashboards, func(dashboard *bizmodel.Dashboard) *adminapi.SelectItem {
		return d.ToSelect(dashboard)
	})
}

func (l *listDashboardRequestBuilder) ToBo() *bo.ListDashboardParams {
	if types.IsNil(l) || types.IsNil(l.ListDashboardRequest) {
		return nil
	}

	return &bo.ListDashboardParams{
		Page:    types.NewPagination(l.GetPagination()),
		Keyword: l.GetKeyword(),
		Status:  vobj.Status(l.GetStatus()),
	}
}

func (d *deleteDashboardRequestBuilder) ToBo() *bo.DeleteDashboardParams {
	if types.IsNil(d) || types.IsNil(d.DeleteDashboardRequest) {
		return nil
	}

	return &bo.DeleteDashboardParams{
		ID:     d.GetId(),
		Status: vobj.StatusDisable,
	}
}

func (u *updateDashboardRequestBuilder) ToBo() *bo.UpdateDashboardParams {
	if types.IsNil(u) || types.IsNil(u.UpdateDashboardRequest) {
		return nil
	}

	return &bo.UpdateDashboardParams{
		ID:             u.GetId(),
		Name:           u.GetTitle(),
		Remark:         u.GetRemark(),
		Color:          u.GetColor(),
		Charts:         NewParamsBuild().WithContext(u.ctx).RealtimeAlarmModuleBuilder().BoChartBuilder().ToBos(u.GetCharts()),
		StrategyGroups: u.GetStrategyGroups(),
	}
}

func (c *createDashboardRequestBuilder) ToBo() *bo.AddDashboardParams {
	if types.IsNil(c) || types.IsNil(c.CreateDashboardRequest) {
		return nil
	}

	return &bo.AddDashboardParams{
		Name:           c.GetTitle(),
		Remark:         c.GetRemark(),
		Color:          c.GetColor(),
		Charts:         NewParamsBuild().WithContext(c.ctx).RealtimeAlarmModuleBuilder().BoChartBuilder().ToBos(c.GetCharts()),
		StrategyGroups: c.GetStrategyGroups(),
	}
}

func (d *doAlarmPageSelfBuilder) ToAPI(self *bizmodel.AlarmPageSelf) *adminapi.SelfAlarmPageItem {
	if types.IsNil(d) || types.IsNil(self) || types.IsNil(self.AlarmPage) {
		return nil
	}

	alarmPageInfo := self.AlarmPage
	return &adminapi.SelfAlarmPageItem{
		Id:           self.ID,
		Name:         alarmPageInfo.GetName(),
		ColorType:    alarmPageInfo.GetColorType(),
		CssClass:     alarmPageInfo.GetCSSClass(),
		Value:        alarmPageInfo.GetValue(),
		Icon:         alarmPageInfo.GetIcon(),
		ImageUrl:     alarmPageInfo.GetImageURL(),
		LanguageCode: alarmPageInfo.GetLanguageCode().String(),
		Remark:       alarmPageInfo.GetRemark(),
	}
}

func (d *doAlarmPageSelfBuilder) ToAPIs(selves []*bizmodel.AlarmPageSelf) []*adminapi.SelfAlarmPageItem {
	if types.IsNil(d) || types.IsNil(selves) {
		return nil
	}

	return types.SliceTo(selves, func(self *bizmodel.AlarmPageSelf) *adminapi.SelfAlarmPageItem {
		return d.ToAPI(self)
	})
}

func (d *doRealtimeAlarmBuilder) ToAPI(alarm *alarmmodel.RealtimeAlarm) *adminapi.RealtimeAlarmItem {
	if types.IsNil(d) || types.IsNil(alarm) {
		return nil
	}

	resItem := &adminapi.RealtimeAlarmItem{
		Id:          alarm.ID,
		StartsAt:    alarm.StartsAt,
		EndsAt:      alarm.EndsAt,
		Status:      api.AlertStatus(alarm.Status),
		Summary:     alarm.Summary,
		Description: alarm.Description,
		Expr:        alarm.Expr,
		Fingerprint: alarm.Fingerprint,
	}
	details := alarm.RealtimeDetails
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

func (d *doRealtimeAlarmBuilder) ToAPIs(alarms []*alarmmodel.RealtimeAlarm) []*adminapi.RealtimeAlarmItem {
	if types.IsNil(d) || types.IsNil(alarms) {
		return nil
	}

	return types.SliceTo(alarms, func(alarm *alarmmodel.RealtimeAlarm) *adminapi.RealtimeAlarmItem {
		return d.ToAPI(alarm)
	})
}

func (l *listAlarmRequestBuilder) ToBo() *bo.GetRealTimeAlarmsParams {
	if types.IsNil(l) || types.IsNil(l.ListAlarmRequest) {
		return nil
	}

	return &bo.GetRealTimeAlarmsParams{
		Pagination:      types.NewPagination(l.GetPagination()),
		EventAtStart:    l.GetEventAtStart(),
		EventAtEnd:      l.GetEventAtEnd(),
		ResolvedAtStart: l.GetRecoverAtStart(),
		ResolvedAtEnd:   l.GetRecoverAtEnd(),
		AlarmLevels:     l.GetAlarmLevels(),
		AlarmStatuses:   types.SliceTo(l.GetAlarmStatuses(), func(status api.AlertStatus) vobj.AlertStatus { return vobj.AlertStatus(status) }),
		Keyword:         l.GetKeyword(),
		AlarmPageID:     l.GetAlarmPage(),
		MyAlarm:         l.GetMyAlarm(),
	}
}

func (g *getAlarmRequestBuilder) ToBo() *bo.GetRealTimeAlarmParams {
	if types.IsNil(g) || types.IsNil(g.GetAlarmRequest) {
		return nil
	}

	return &bo.GetRealTimeAlarmParams{
		RealtimeAlarmID: g.GetId(),
	}
}

func (r *realtimeAlarmModuleBuilder) WithGetAlarmRequest(request *realtimeapi.GetAlarmRequest) IGetAlarmRequestBuilder {
	return &getAlarmRequestBuilder{ctx: r.ctx, GetAlarmRequest: request}
}

func (r *realtimeAlarmModuleBuilder) WithListAlarmRequest(request *realtimeapi.ListAlarmRequest) IListAlarmRequestBuilder {
	return &listAlarmRequestBuilder{ctx: r.ctx, ListAlarmRequest: request}
}

func (r *realtimeAlarmModuleBuilder) DoRealtimeAlarmBuilder() IDoRealtimeAlarmBuilder {
	return &doRealtimeAlarmBuilder{ctx: r.ctx}
}

func (r *realtimeAlarmModuleBuilder) DoAlarmPageSelfBuilder() IDoAlarmPageSelfBuilder {
	return &doAlarmPageSelfBuilder{ctx: r.ctx}
}

func (r *realtimeAlarmModuleBuilder) WithCreateDashboardRequest(request *realtimeapi.CreateDashboardRequest) ICreateDashboardRequestBuilder {
	return &createDashboardRequestBuilder{ctx: r.ctx, CreateDashboardRequest: request}
}

func (r *realtimeAlarmModuleBuilder) WithUpdateDashboardRequest(request *realtimeapi.UpdateDashboardRequest) IUpdateDashboardRequestBuilder {
	return &updateDashboardRequestBuilder{ctx: r.ctx, UpdateDashboardRequest: request}
}

func (r *realtimeAlarmModuleBuilder) WithDeleteDashboardRequest(request *realtimeapi.DeleteDashboardRequest) IDeleteDashboardRequestBuilder {
	return &deleteDashboardRequestBuilder{ctx: r.ctx, DeleteDashboardRequest: request}
}

func (r *realtimeAlarmModuleBuilder) WithListDashboardRequest(request *realtimeapi.ListDashboardRequest) IListDashboardRequestBuilder {
	return &listDashboardRequestBuilder{ctx: r.ctx, ListDashboardRequest: request}
}

func (r *realtimeAlarmModuleBuilder) DoDashboardBuilder() IDoDashboardBuilder {
	return &doDashboardBuilder{ctx: r.ctx}
}
