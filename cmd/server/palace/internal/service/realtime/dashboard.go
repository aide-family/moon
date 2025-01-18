package realtime

import (
	"context"

	realtimeapi "github.com/aide-family/moon/api/admin/realtime"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
)

// DashboardService 监控大盘服务
type DashboardService struct {
	realtimeapi.UnimplementedDashboardServer

	dashboardBiz *biz.DashboardBiz
}

// NewDashboardService 创建监控大盘服务
func NewDashboardService(dashboardBiz *biz.DashboardBiz) *DashboardService {
	return &DashboardService{
		dashboardBiz: dashboardBiz,
	}
}

// CreateDashboard 创建监控大盘
func (s *DashboardService) CreateDashboard(ctx context.Context, req *realtimeapi.CreateDashboardRequest) (*realtimeapi.CreateDashboardReply, error) {
	params := builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().WithCreateDashboardRequest(req).ToBo()
	if err := s.dashboardBiz.CreateDashboard(ctx, params); err != nil {
		return nil, err
	}
	return &realtimeapi.CreateDashboardReply{}, nil
}

// UpdateDashboard 更新监控大盘
func (s *DashboardService) UpdateDashboard(ctx context.Context, req *realtimeapi.UpdateDashboardRequest) (*realtimeapi.UpdateDashboardReply, error) {
	params := builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().WithUpdateDashboardRequest(req).ToBo()
	if err := s.dashboardBiz.UpdateDashboard(ctx, params); err != nil {
		return nil, err
	}
	return &realtimeapi.UpdateDashboardReply{}, nil
}

// DeleteDashboard 删除监控大盘
func (s *DashboardService) DeleteDashboard(ctx context.Context, req *realtimeapi.DeleteDashboardRequest) (*realtimeapi.DeleteDashboardReply, error) {
	params := builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().WithDeleteDashboardRequest(req).ToBo()
	if err := s.dashboardBiz.DeleteDashboard(ctx, params); err != nil {
		return nil, err
	}
	return &realtimeapi.DeleteDashboardReply{}, nil
}

// GetDashboard 获取监控大盘
func (s *DashboardService) GetDashboard(ctx context.Context, req *realtimeapi.GetDashboardRequest) (*realtimeapi.GetDashboardReply, error) {
	params := builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().WithGetDashboardRequest(req).ToBo()
	detail, err := s.dashboardBiz.GetDashboard(ctx, params)
	if err != nil {
		return nil, err
	}
	return &realtimeapi.GetDashboardReply{
		Detail: builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().DoDashboardBuilder().ToAPI(detail),
	}, nil
}

// ListDashboard 获取监控大盘列表
func (s *DashboardService) ListDashboard(ctx context.Context, req *realtimeapi.ListDashboardRequest) (*realtimeapi.ListDashboardReply, error) {
	params := builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().WithListDashboardRequest(req).ToBo()
	list, err := s.dashboardBiz.ListDashboard(ctx, params)
	if err != nil {
		return nil, err
	}

	return &realtimeapi.ListDashboardReply{
		List:       builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().DoDashboardBuilder().ToAPIs(list),
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(params.Page),
	}, nil
}

// ListDashboardSelect 获取监控大盘下拉列表
func (s *DashboardService) ListDashboardSelect(ctx context.Context, req *realtimeapi.ListDashboardRequest) (*realtimeapi.ListDashboardSelectReply, error) {
	params := builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().WithListDashboardRequest(req).ToBo()
	list, err := s.dashboardBiz.ListDashboard(ctx, params)
	if err != nil {
		return nil, err
	}
	return &realtimeapi.ListDashboardSelectReply{
		List:       builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().DoDashboardBuilder().ToSelects(list),
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(params.Page),
	}, nil
}

// BatchUpdateDashboardStatus 批量更新监控大盘状态
func (s *DashboardService) BatchUpdateDashboardStatus(ctx context.Context, req *realtimeapi.BatchUpdateDashboardStatusRequest) (*realtimeapi.BatchUpdateDashboardStatusReply, error) {
	params := builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().WithBatchUpdateDashboardStatusRequest(req).ToBo()
	if err := s.dashboardBiz.BatchUpdateDashboardStatus(ctx, params); err != nil {
		return nil, err
	}
	return &realtimeapi.BatchUpdateDashboardStatusReply{}, nil
}

// AddChart 添加图表
func (s *DashboardService) AddChart(ctx context.Context, req *realtimeapi.AddChartRequest) (*realtimeapi.AddChartReply, error) {
	params := builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().WithAddChartRequest(req).ToBo()
	if err := s.dashboardBiz.AddChart(ctx, params); err != nil {
		return nil, err
	}
	return &realtimeapi.AddChartReply{}, nil
}

// UpdateChart 更新图表
func (s *DashboardService) UpdateChart(ctx context.Context, req *realtimeapi.UpdateChartRequest) (*realtimeapi.UpdateChartReply, error) {
	params := builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().WithUpdateChartRequest(req).ToBo()
	if err := s.dashboardBiz.UpdateChart(ctx, params); err != nil {
		return nil, err
	}
	return &realtimeapi.UpdateChartReply{}, nil
}

// DeleteChart 删除图表
func (s *DashboardService) DeleteChart(ctx context.Context, req *realtimeapi.DeleteChartRequest) (*realtimeapi.DeleteChartReply, error) {
	params := builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().WithDeleteChartRequest(req).ToBo()
	if err := s.dashboardBiz.DeleteChart(ctx, params); err != nil {
		return nil, err
	}
	return &realtimeapi.DeleteChartReply{}, nil
}

// GetChart 获取图表
func (s *DashboardService) GetChart(ctx context.Context, req *realtimeapi.GetChartRequest) (*realtimeapi.GetChartReply, error) {
	params := builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().WithGetChartRequest(req).ToBo()
	detail, err := s.dashboardBiz.GetChart(ctx, params)
	if err != nil {
		return nil, err
	}
	return &realtimeapi.GetChartReply{
		Detail: builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().DoChartBuilder().ToAPI(detail),
	}, nil
}

// ListChart 获取图表列表
func (s *DashboardService) ListChart(ctx context.Context, req *realtimeapi.ListChartRequest) (*realtimeapi.ListChartReply, error) {
	params := builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().WithListChartRequest(req).ToBo()
	list, err := s.dashboardBiz.ListChart(ctx, params)
	if err != nil {
		return nil, err
	}
	return &realtimeapi.ListChartReply{
		List:       builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().DoChartBuilder().ToAPIs(list),
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(params.Page),
	}, nil
}

// BatchUpdateChartStatus 批量更新图表状态
func (s *DashboardService) BatchUpdateChartStatus(ctx context.Context, req *realtimeapi.BatchUpdateChartStatusRequest) (*realtimeapi.BatchUpdateChartStatusReply, error) {
	params := builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().WithBatchUpdateChartStatusRequest(req).ToBo()
	if err := s.dashboardBiz.BatchUpdateChartStatus(ctx, params); err != nil {
		return nil, err
	}
	return &realtimeapi.BatchUpdateChartStatusReply{}, nil
}

// BatchUpdateChartSort 批量更新图表排序
func (s *DashboardService) BatchUpdateChartSort(ctx context.Context, req *realtimeapi.BatchUpdateChartSortRequest) (*realtimeapi.BatchUpdateChartSortReply, error) {
	if err := s.dashboardBiz.BatchUpdateChartSort(ctx, req.GetDashboardId(), req.GetIds()); err != nil {
		return nil, err
	}
	return &realtimeapi.BatchUpdateChartSortReply{}, nil
}

// ListSelfDashboard 获取个人仪表板列表
func (s *DashboardService) ListSelfDashboard(ctx context.Context, req *realtimeapi.ListSelfDashboardRequest) (*realtimeapi.ListSelfDashboardReply, error) {
	list, err := s.dashboardBiz.ListSelfDashboard(ctx)
	if err != nil {
		return nil, err
	}
	return &realtimeapi.ListSelfDashboardReply{
		List: builder.NewParamsBuild(ctx).RealtimeAlarmModuleBuilder().DoDashboardBuilder().ToAPIs(list),
	}, nil
}

// UpdateSelfDashboard 更新个人仪表板
func (s *DashboardService) UpdateSelfDashboard(ctx context.Context, req *realtimeapi.UpdateSelfDashboardRequest) (*realtimeapi.UpdateSelfDashboardReply, error) {
	if err := s.dashboardBiz.UpdateSelfDashboard(ctx, req.GetIds()); err != nil {
		return nil, err
	}
	return &realtimeapi.UpdateSelfDashboardReply{}, nil
}
