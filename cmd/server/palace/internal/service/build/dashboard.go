package build

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	realtimeapi "github.com/aide-family/moon/api/admin/realtime"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// NewDashboardModuleBuilder 创建仪表盘模块构造器
func NewDashboardModuleBuilder(ctx context.Context) DashboardModuleBuilder {
	return &dashboardModuleBuilder{ctx: ctx}
}

func newAPIChartItemBuilder(ctx context.Context, item *adminapi.ChartItem) APIChartItemBuilder {
	return &apiChartItemBuilder{ctx: ctx, item: item}
}

func newAPIsChartItemBuilder(ctx context.Context, items []*adminapi.ChartItem) APIsChartItemBuilder {
	return &apisChartItemBuilder{ctx: ctx, items: items}
}

func newDoChartsBuilder(ctx context.Context, item *bizmodel.DashboardChart) DoChartBuilder {
	return &doChartBuilder{ctx: ctx, item: item}
}

func newDosChartsBuilder(ctx context.Context, items []*bizmodel.DashboardChart) DosChartBuilder {
	return &dosChartBuilder{ctx: ctx, items: items}
}

func newDoDashboardBuilder(ctx context.Context, item *bizmodel.Dashboard) DoDashboardBuilder {
	return &doDashboardBuilder{ctx: ctx, item: item}
}

func newDosDashboardBuilder(ctx context.Context, items []*bizmodel.Dashboard) DosDashboardBuilder {
	return &dosDashboardBuilder{ctx: ctx, items: items}
}

func newBoAddDashboardParamsBuilder(ctx context.Context, params *bo.AddDashboardParams) BoAddDashboardParamsBuilder {
	return &boAddDashboardParamsBuilder{ctx: ctx, params: params}
}

func newBoChartItemBuilder(ctx context.Context, item *bo.ChartItem) BoChartItemBuilder {
	return &boChartItemBuilder{ctx: ctx, item: item}
}

func newBosChartItemBuilder(ctx context.Context, items []*bo.ChartItem) BosChartItemBuilder {
	return &bosChartItemBuilder{ctx: ctx, items: items}
}

func newBoUpdateDashboardParamsBuilder(ctx context.Context, params *bo.UpdateDashboardParams) BoUpdateDashboardParamsBuilder {
	return &boUpdateDashboardParamsBuilder{ctx: ctx, params: params}
}

func newAPIDeleteDashboardParamsBuilder(ctx context.Context, params *realtimeapi.DeleteDashboardRequest) APIDeleteDashboardParamsBuilder {
	return &apiDeleteDashboardParamsBuilder{ctx: ctx, params: params}
}

func newAPIDashboardListParamsBuilder(ctx context.Context, params *realtimeapi.ListDashboardRequest) APIListDashboardParams {
	return &apiListDashboardParams{ctx: ctx, params: params}
}

func newAPIDashboardSelectParamsBuilder(ctx context.Context, params *realtimeapi.ListDashboardSelectRequest) APIListDashboardParams {
	return &apiListDashboardSelectParams{ctx: ctx, params: params}
}

func newAPIUpdateDashboardParamsBuilder(ctx context.Context, params *realtimeapi.UpdateDashboardRequest) APIUpdateDashboardParamsBuilder {
	return &apiUpdateDashboardParamsBuilder{ctx: ctx, params: params}
}

func newAPIAddDashboardParamsBuilder(ctx context.Context, params *realtimeapi.CreateDashboardRequest) APIAddDashboardParamsBuilder {
	return &apiAddDashboardParamsBuilder{ctx: ctx, params: params}
}

type (
	// APIChartItemBuilder bo 图表构造器
	APIChartItemBuilder interface {
		ToBo() *bo.ChartItem
		WithDashboardID(uint32) APIChartItemBuilder
	}

	// APIsChartItemBuilder 图表构造器
	APIsChartItemBuilder interface {
		ToBos() []*bo.ChartItem
		WithDashboardID(uint32) APIsChartItemBuilder
	}

	// BoChartItemBuilder 图表构造器
	BoChartItemBuilder interface {
		WithDashboardID(uint32) BoChartItemBuilder
		ToDo() *bizmodel.DashboardChart
	}

	// BosChartItemBuilder 图表构造器
	BosChartItemBuilder interface {
		WithDashboardID(uint32) BosChartItemBuilder
		ToDos() []*bizmodel.DashboardChart
	}

	// APIAddDashboardParamsBuilder 添加仪表盘参数构造器
	APIAddDashboardParamsBuilder interface {
		ToBo() *bo.AddDashboardParams
	}

	// APIUpdateDashboardParamsBuilder 更新仪表盘参数构造器
	APIUpdateDashboardParamsBuilder interface {
		ToBo() *bo.UpdateDashboardParams
	}

	// APIDeleteDashboardParamsBuilder 删除仪表盘参数构造器
	APIDeleteDashboardParamsBuilder interface {
		ToBo() *bo.DeleteDashboardParams
	}

	// APIListDashboardParams 列表构造器
	APIListDashboardParams interface {
		ToBo() *bo.ListDashboardParams
	}

	// DoDashboardBuilder 仪表盘构造器
	DoDashboardBuilder interface {
		ToAPI() *adminapi.DashboardItem
		ToSelect() *adminapi.SelectItem
	}

	// DosDashboardBuilder 仪表盘列表构造器
	DosDashboardBuilder interface {
		ToAPIs() []*adminapi.DashboardItem
		ToSelects() []*adminapi.SelectItem
	}

	// BoAddDashboardParamsBuilder 添加仪表盘参数构造器
	BoAddDashboardParamsBuilder interface {
		ToDo() *bizmodel.Dashboard
		ToDoCharts() []*bizmodel.DashboardChart
		ToDoStrategyGroups() []*bizmodel.StrategyGroup
		WithDashboardID(uint32) BoAddDashboardParamsBuilder
	}

	// BoUpdateDashboardParamsBuilder 更新仪表盘参数构造器
	BoUpdateDashboardParamsBuilder interface {
		ToDo() *bizmodel.Dashboard
		ToDoCharts() []*bizmodel.DashboardChart
		ToDoStrategyGroups() []*bizmodel.StrategyGroup
		WithDashboardID(uint32) BoUpdateDashboardParamsBuilder
	}

	// DosChartBuilder 图表构造器
	DosChartBuilder interface {
		ToAPIs() []*adminapi.ChartItem
		ToSelects() []*adminapi.SelectItem
	}

	// DoChartBuilder 图表构造器
	DoChartBuilder interface {
		ToAPI() *adminapi.ChartItem
		ToSelect() *adminapi.SelectItem
	}

	// DashboardModuleBuilder 仪表盘模块构造器
	DashboardModuleBuilder interface {
		WithAPIChart(*adminapi.ChartItem) APIChartItemBuilder
		WithAPIsChart(items []*adminapi.ChartItem) APIsChartItemBuilder
		WithAPIAddDashboardParams(*realtimeapi.CreateDashboardRequest) APIAddDashboardParamsBuilder
		WithAPIUpdateDashboardParams(*realtimeapi.UpdateDashboardRequest) APIUpdateDashboardParamsBuilder
		WithAPIDeleteDashboardParams(*realtimeapi.DeleteDashboardRequest) APIDeleteDashboardParamsBuilder
		WithAPIQueryDashboardListParams(*realtimeapi.ListDashboardRequest) APIListDashboardParams
		WithAPIQueryDashboardSelectParams(*realtimeapi.ListDashboardSelectRequest) APIListDashboardParams
		WithDoDashboard(*bizmodel.Dashboard) DoDashboardBuilder
		WithDoDashboardList([]*bizmodel.Dashboard) DosDashboardBuilder
		WithBoAddDashboardParams(*bo.AddDashboardParams) BoAddDashboardParamsBuilder
		WithBoUpdateDashboardParams(*bo.UpdateDashboardParams) BoUpdateDashboardParamsBuilder
		WithBoChart(*bo.ChartItem) BoChartItemBuilder
		WithBosChart([]*bo.ChartItem) BosChartItemBuilder
		WithDosChart([]*bizmodel.DashboardChart) DosChartBuilder
		WithDoChart(*bizmodel.DashboardChart) DoChartBuilder
	}

	dashboardModuleBuilder struct {
		ctx context.Context
	}

	apiChartItemBuilder struct {
		ctx         context.Context
		item        *adminapi.ChartItem
		dashboardID uint32
	}

	apisChartItemBuilder struct {
		ctx         context.Context
		items       []*adminapi.ChartItem
		dashboardID uint32
	}

	apiAddDashboardParamsBuilder struct {
		ctx    context.Context
		params *realtimeapi.CreateDashboardRequest
	}

	apiUpdateDashboardParamsBuilder struct {
		ctx    context.Context
		params *realtimeapi.UpdateDashboardRequest
	}

	apiDeleteDashboardParamsBuilder struct {
		ctx    context.Context
		params *realtimeapi.DeleteDashboardRequest
	}

	apiListDashboardParams struct {
		ctx    context.Context
		params *realtimeapi.ListDashboardRequest
	}

	apiListDashboardSelectParams struct {
		ctx    context.Context
		params *realtimeapi.ListDashboardSelectRequest
	}

	doDashboardBuilder struct {
		ctx  context.Context
		item *bizmodel.Dashboard
	}

	dosDashboardBuilder struct {
		ctx   context.Context
		items []*bizmodel.Dashboard
	}

	dosChartBuilder struct {
		ctx   context.Context
		items []*bizmodel.DashboardChart
	}

	doChartBuilder struct {
		ctx  context.Context
		item *bizmodel.DashboardChart
	}

	boAddDashboardParamsBuilder struct {
		ctx         context.Context
		params      *bo.AddDashboardParams
		dashboardID uint32
	}

	boUpdateDashboardParamsBuilder struct {
		ctx         context.Context
		params      *bo.UpdateDashboardParams
		dashboardID uint32
	}

	boChartItemBuilder struct {
		ctx         context.Context
		item        *bo.ChartItem
		dashboardID uint32
	}

	bosChartItemBuilder struct {
		ctx         context.Context
		items       []*bo.ChartItem
		dashboardID uint32
	}
)

func (b *boUpdateDashboardParamsBuilder) ToDo() *bizmodel.Dashboard {
	if types.IsNil(b) || types.IsNil(b.params) {
		return nil
	}
	params := b.params
	return &bizmodel.Dashboard{
		AllFieldModel: model.AllFieldModel{
			ID: params.ID,
		},
		Name:           params.Name,
		Remark:         params.Remark,
		Status:         params.Status,
		Color:          b.params.Color,
		Charts:         b.ToDoCharts(),
		StrategyGroups: b.ToDoStrategyGroups(),
	}
}

func (b *boUpdateDashboardParamsBuilder) ToDoCharts() []*bizmodel.DashboardChart {
	if types.IsNil(b) || types.IsNil(b.params) {
		return nil
	}
	return newBosChartItemBuilder(b.ctx, b.params.Charts).WithDashboardID(b.dashboardID).ToDos()
}

func (b *boUpdateDashboardParamsBuilder) ToDoStrategyGroups() []*bizmodel.StrategyGroup {
	if types.IsNil(b) || types.IsNil(b.params) {
		return nil
	}
	return types.SliceTo(b.params.StrategyGroups, func(item uint32) *bizmodel.StrategyGroup {
		return &bizmodel.StrategyGroup{AllFieldModel: model.AllFieldModel{ID: item}}
	})
}

func (b *boUpdateDashboardParamsBuilder) WithDashboardID(dashboardID uint32) BoUpdateDashboardParamsBuilder {
	if types.IsNil(b) || types.IsNil(b.params) {
		return newBoUpdateDashboardParamsBuilder(b.ctx, nil).WithDashboardID(dashboardID)
	}
	b.dashboardID = dashboardID
	return b
}

func (b *bosChartItemBuilder) WithDashboardID(dashboardID uint32) BosChartItemBuilder {
	if types.IsNil(b) || types.IsNil(b.items) {
		return newBosChartItemBuilder(b.ctx, nil).WithDashboardID(dashboardID)
	}
	b.dashboardID = dashboardID
	return b
}

func (b *bosChartItemBuilder) ToDos() []*bizmodel.DashboardChart {
	if types.IsNil(b) || types.IsNil(b.items) {
		return nil
	}
	items := b.items
	return types.SliceTo(items, func(item *bo.ChartItem) *bizmodel.DashboardChart {
		return newBoChartItemBuilder(b.ctx, item).WithDashboardID(b.dashboardID).ToDo()
	})
}

func (b *boChartItemBuilder) WithDashboardID(dashboardID uint32) BoChartItemBuilder {
	if types.IsNil(b) || types.IsNil(b.item) {
		return newBoChartItemBuilder(b.ctx, nil).WithDashboardID(dashboardID)
	}
	b.dashboardID = dashboardID
	return b
}

func (b *boChartItemBuilder) ToDo() *bizmodel.DashboardChart {
	if types.IsNil(b) || types.IsNil(b.item) {
		return nil
	}
	item := b.item
	return &bizmodel.DashboardChart{
		AllFieldModel: model.AllFieldModel{ID: item.ID},
		Name:          item.Name,
		Status:        item.Status,
		Remark:        item.Remark,
		URL:           item.URL,
		DashboardID:   b.dashboardID,
		ChartType:     item.ChartType,
		Width:         item.Width,
		Height:        item.Height,
	}
}

func (b *boAddDashboardParamsBuilder) ToDo() *bizmodel.Dashboard {
	if types.IsNil(b) || types.IsNil(b.params) {
		return nil
	}
	params := b.params
	return &bizmodel.Dashboard{
		Name:           params.Name,
		Status:         vobj.StatusEnable,
		Remark:         params.Remark,
		Color:          params.Color,
		Charts:         b.ToDoCharts(),
		StrategyGroups: b.ToDoStrategyGroups(),
	}
}

func (b *boAddDashboardParamsBuilder) ToDoCharts() []*bizmodel.DashboardChart {
	if types.IsNil(b) || types.IsNil(b.params) {
		return nil
	}
	return newBosChartItemBuilder(b.ctx, b.params.Charts).ToDos()
}

func (b *boAddDashboardParamsBuilder) ToDoStrategyGroups() []*bizmodel.StrategyGroup {
	if types.IsNil(b) || types.IsNil(b.params) {
		return nil
	}
	return types.SliceTo(b.params.StrategyGroups, func(item uint32) *bizmodel.StrategyGroup {
		return &bizmodel.StrategyGroup{
			AllFieldModel: model.AllFieldModel{ID: item},
		}
	})
}

func (b *boAddDashboardParamsBuilder) WithDashboardID(dashboardID uint32) BoAddDashboardParamsBuilder {
	if types.IsNil(b) {
		return newBoAddDashboardParamsBuilder(context.TODO(), nil).WithDashboardID(dashboardID)
	}
	b.dashboardID = dashboardID
	return b
}

func (d *dosDashboardBuilder) ToAPIs() []*adminapi.DashboardItem {
	if types.IsNil(d) || types.IsNil(d.items) {
		return nil
	}
	return types.SliceTo(d.items, func(item *bizmodel.Dashboard) *adminapi.DashboardItem {
		return newDoDashboardBuilder(d.ctx, item).ToAPI()
	})
}

func (d *dosDashboardBuilder) ToSelects() []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(d.items) {
		return nil
	}
	return types.SliceTo(d.items, func(item *bizmodel.Dashboard) *adminapi.SelectItem {
		return newDoDashboardBuilder(d.ctx, item).ToSelect()
	})
}

func (d *dosChartBuilder) ToAPIs() []*adminapi.ChartItem {
	if types.IsNil(d) || types.IsNil(d.items) {
		return nil
	}
	items := d.items
	return types.SliceTo(items, func(item *bizmodel.DashboardChart) *adminapi.ChartItem {
		return newDoChartsBuilder(d.ctx, item).ToAPI()
	})
}

func (d *dosChartBuilder) ToSelects() []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(d.items) {
		return nil
	}
	items := d.items
	return types.SliceTo(items, func(item *bizmodel.DashboardChart) *adminapi.SelectItem {
		return newDoChartsBuilder(d.ctx, item).ToSelect()
	})
}

func (d *doChartBuilder) ToAPI() *adminapi.ChartItem {
	if types.IsNil(d) || types.IsNil(d.item) {
		return nil
	}

	item := d.item
	return &adminapi.ChartItem{
		Id:        item.ID,
		Title:     item.Name,
		Remark:    item.Remark,
		Url:       item.URL,
		Status:    api.Status(item.Status),
		ChartType: api.ChartType(item.ChartType),
		Width:     item.Width,
		Height:    item.Height,
	}
}

func (d *doChartBuilder) ToSelect() *adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(d.item) {
		return nil
	}
	item := d.item
	return &adminapi.SelectItem{
		Value:    item.ID,
		Label:    item.Name,
		Children: nil,
		Disabled: item.DeletedAt > 0 || !item.Status.IsEnable(),
		Extend:   nil,
	}
}

func (d *dashboardModuleBuilder) WithDosChart(charts []*bizmodel.DashboardChart) DosChartBuilder {
	return &dosChartBuilder{ctx: d.ctx, items: charts}
}

func (d *dashboardModuleBuilder) WithDoChart(chart *bizmodel.DashboardChart) DoChartBuilder {
	return &doChartBuilder{ctx: d.ctx, item: chart}
}

func (d *doDashboardBuilder) ToAPI() *adminapi.DashboardItem {
	if types.IsNil(d) || types.IsNil(d.item) {
		return nil
	}
	item := d.item
	return &adminapi.DashboardItem{
		Id:        item.ID,
		Title:     item.Name,
		Remark:    item.Remark,
		CreatedAt: item.CreatedAt.String(),
		UpdatedAt: item.UpdatedAt.String(),
		DeletedAt: types.NewTimeByUnix(int64(item.DeletedAt)).String(),
		Color:     item.Color,
		Charts:    newDosChartsBuilder(d.ctx, item.Charts).ToAPIs(),
		Status:    api.Status(item.Status),
		Groups:    nil, // TODO 等builder重构完一起替换
	}
}

func (d *doDashboardBuilder) ToSelect() *adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(d.item) {
		return nil
	}
	item := d.item
	return &adminapi.SelectItem{
		Value:    item.ID,
		Label:    item.Name,
		Children: nil,
		Disabled: item.DeletedAt > 9 || !item.Status.IsEnable(),
		Extend:   nil,
	}
}

func (a *apiListDashboardSelectParams) ToBo() *bo.ListDashboardParams {
	if types.IsNil(a) || types.IsNil(a.params) {
		return nil
	}
	params := a.params
	return &bo.ListDashboardParams{
		Page:    types.NewPagination(params.GetPagination()),
		Keyword: params.GetKeyword(),
		Status:  vobj.Status(params.GetStatus()),
	}
}

func (a *apiListDashboardParams) ToBo() *bo.ListDashboardParams {
	if types.IsNil(a) || types.IsNil(a.params) {
		return nil
	}
	params := a.params
	return &bo.ListDashboardParams{
		Page:    types.NewPagination(params.GetPagination()),
		Keyword: params.GetKeyword(),
		Status:  vobj.Status(params.GetStatus()),
	}
}

func (a *apiDeleteDashboardParamsBuilder) ToBo() *bo.DeleteDashboardParams {
	if types.IsNil(a) || types.IsNil(a.params) {
		return nil
	}
	params := a.params
	return &bo.DeleteDashboardParams{
		ID:     params.GetId(),
		Status: vobj.StatusDisable, // 只允许删除禁用状态的仪表盘
	}
}

func (a *apiUpdateDashboardParamsBuilder) ToBo() *bo.UpdateDashboardParams {
	if types.IsNil(a) || types.IsNil(a.params) {
		return nil
	}
	params := a.params
	return &bo.UpdateDashboardParams{
		ID:             params.GetId(),
		Name:           params.GetTitle(),
		Remark:         params.GetRemark(),
		Color:          params.GetColor(),
		Charts:         newAPIsChartItemBuilder(a.ctx, params.GetCharts()).WithDashboardID(params.GetId()).ToBos(),
		StrategyGroups: params.GetStrategyGroups(),
	}
}

func (a *apiAddDashboardParamsBuilder) ToBo() *bo.AddDashboardParams {
	if types.IsNil(a) || types.IsNil(a.params) {
		return nil
	}
	params := a.params
	return &bo.AddDashboardParams{
		Name:           params.GetTitle(),
		Remark:         params.GetRemark(),
		Color:          params.GetColor(),
		Charts:         newAPIsChartItemBuilder(a.ctx, params.GetCharts()).ToBos(),
		StrategyGroups: params.GetStrategyGroups(),
	}
}

func (a *apisChartItemBuilder) ToBos() []*bo.ChartItem {
	if types.IsNil(a) {
		return nil
	}
	return types.SliceTo(a.items, func(item *adminapi.ChartItem) *bo.ChartItem {
		return newAPIChartItemBuilder(a.ctx, item).WithDashboardID(a.dashboardID).ToBo()
	})
}

func (a *apisChartItemBuilder) WithDashboardID(dashboardID uint32) APIsChartItemBuilder {
	if types.IsNil(a) {
		return &apisChartItemBuilder{ctx: context.TODO()}
	}
	a.dashboardID = dashboardID
	return a
}

func (a *apiChartItemBuilder) ToBo() *bo.ChartItem {
	if types.IsNil(a) || types.IsNil(a.item) {
		return nil
	}
	item := a.item
	return &bo.ChartItem{
		ID:          item.GetId(),
		Name:        item.GetTitle(),
		Remark:      item.GetRemark(),
		URL:         item.GetUrl(),
		Status:      vobj.Status(item.GetStatus()),
		Height:      item.GetHeight(),
		Width:       item.GetWidth(),
		ChartType:   vobj.DashboardChartType(a.item.ChartType),
		DashboardID: a.dashboardID,
	}
}

func (a *apiChartItemBuilder) WithDashboardID(dashboardID uint32) APIChartItemBuilder {
	if types.IsNil(a) {
		return &apiChartItemBuilder{ctx: context.TODO()}
	}
	a.dashboardID = dashboardID
	return a
}

func (d *dashboardModuleBuilder) WithAPIChart(item *adminapi.ChartItem) APIChartItemBuilder {
	return newAPIChartItemBuilder(d.ctx, item)
}

func (d *dashboardModuleBuilder) WithAPIsChart(items []*adminapi.ChartItem) APIsChartItemBuilder {
	return newAPIsChartItemBuilder(d.ctx, items)
}

func (d *dashboardModuleBuilder) WithAPIAddDashboardParams(request *realtimeapi.CreateDashboardRequest) APIAddDashboardParamsBuilder {
	return newAPIAddDashboardParamsBuilder(d.ctx, request)
}

func (d *dashboardModuleBuilder) WithAPIUpdateDashboardParams(request *realtimeapi.UpdateDashboardRequest) APIUpdateDashboardParamsBuilder {
	return newAPIUpdateDashboardParamsBuilder(d.ctx, request)
}

func (d *dashboardModuleBuilder) WithAPIDeleteDashboardParams(request *realtimeapi.DeleteDashboardRequest) APIDeleteDashboardParamsBuilder {
	return newAPIDeleteDashboardParamsBuilder(d.ctx, request)
}

func (d *dashboardModuleBuilder) WithAPIQueryDashboardListParams(request *realtimeapi.ListDashboardRequest) APIListDashboardParams {
	return newAPIDashboardListParamsBuilder(d.ctx, request)
}

func (d *dashboardModuleBuilder) WithAPIQueryDashboardSelectParams(request *realtimeapi.ListDashboardSelectRequest) APIListDashboardParams {
	return newAPIDashboardSelectParamsBuilder(d.ctx, request)
}

func (d *dashboardModuleBuilder) WithDoDashboard(dashboard *bizmodel.Dashboard) DoDashboardBuilder {
	return newDoDashboardBuilder(d.ctx, dashboard)
}

func (d *dashboardModuleBuilder) WithDoDashboardList(dashboards []*bizmodel.Dashboard) DosDashboardBuilder {
	return newDosDashboardBuilder(d.ctx, dashboards)
}

func (d *dashboardModuleBuilder) WithBoAddDashboardParams(params *bo.AddDashboardParams) BoAddDashboardParamsBuilder {
	return newBoAddDashboardParamsBuilder(d.ctx, params)
}

func (d *dashboardModuleBuilder) WithBoUpdateDashboardParams(params *bo.UpdateDashboardParams) BoUpdateDashboardParamsBuilder {
	return newBoUpdateDashboardParamsBuilder(d.ctx, params)
}

func (d *dashboardModuleBuilder) WithBoChart(item *bo.ChartItem) BoChartItemBuilder {
	return newBoChartItemBuilder(d.ctx, item)
}

func (d *dashboardModuleBuilder) WithBosChart(items []*bo.ChartItem) BosChartItemBuilder {
	return newBosChartItemBuilder(d.ctx, items)
}
