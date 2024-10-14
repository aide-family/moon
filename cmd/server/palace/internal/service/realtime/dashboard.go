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
	params := builder.NewParamsBuild().WithContext(ctx).RealtimeAlarmModuleBuilder().WithCreateDashboardRequest(req).ToBo()
	if err := s.dashboardBiz.CreateDashboard(ctx, params); err != nil {
		return nil, err
	}
	return &realtimeapi.CreateDashboardReply{}, nil
}

// UpdateDashboard 更新监控大盘
func (s *DashboardService) UpdateDashboard(ctx context.Context, req *realtimeapi.UpdateDashboardRequest) (*realtimeapi.UpdateDashboardReply, error) {
	params := builder.NewParamsBuild().WithContext(ctx).RealtimeAlarmModuleBuilder().WithUpdateDashboardRequest(req).ToBo()
	if err := s.dashboardBiz.UpdateDashboard(ctx, params); err != nil {
		return nil, err
	}
	return &realtimeapi.UpdateDashboardReply{}, nil
}

// DeleteDashboard 删除监控大盘
func (s *DashboardService) DeleteDashboard(ctx context.Context, req *realtimeapi.DeleteDashboardRequest) (*realtimeapi.DeleteDashboardReply, error) {
	params := builder.NewParamsBuild().WithContext(ctx).RealtimeAlarmModuleBuilder().WithDeleteDashboardRequest(req).ToBo()
	if err := s.dashboardBiz.DeleteDashboard(ctx, params); err != nil {
		return nil, err
	}
	return &realtimeapi.DeleteDashboardReply{}, nil
}

// GetDashboard 获取监控大盘
func (s *DashboardService) GetDashboard(ctx context.Context, req *realtimeapi.GetDashboardRequest) (*realtimeapi.GetDashboardReply, error) {
	detail, err := s.dashboardBiz.GetDashboard(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &realtimeapi.GetDashboardReply{
		Detail: builder.NewParamsBuild().WithContext(ctx).RealtimeAlarmModuleBuilder().DoDashboardBuilder().ToAPI(detail),
	}, nil
}

// ListDashboard 获取监控大盘列表
func (s *DashboardService) ListDashboard(ctx context.Context, req *realtimeapi.ListDashboardRequest) (*realtimeapi.ListDashboardReply, error) {
	params := builder.NewParamsBuild().WithContext(ctx).RealtimeAlarmModuleBuilder().WithListDashboardRequest(req).ToBo()
	list, err := s.dashboardBiz.ListDashboard(ctx, params)
	if err != nil {
		return nil, err
	}

	return &realtimeapi.ListDashboardReply{
		List:       builder.NewParamsBuild().WithContext(ctx).RealtimeAlarmModuleBuilder().DoDashboardBuilder().ToAPIs(list),
		Pagination: builder.NewParamsBuild().WithContext(ctx).PaginationModuleBuilder().ToAPI(params.Page),
	}, nil
}

// ListDashboardSelect 获取监控大盘下拉列表
func (s *DashboardService) ListDashboardSelect(ctx context.Context, req *realtimeapi.ListDashboardRequest) (*realtimeapi.ListDashboardSelectReply, error) {
	params := builder.NewParamsBuild().RealtimeAlarmModuleBuilder().WithListDashboardRequest(req).ToBo()
	list, err := s.dashboardBiz.ListDashboard(ctx, params)
	if err != nil {
		return nil, err
	}
	return &realtimeapi.ListDashboardSelectReply{
		List:       builder.NewParamsBuild().RealtimeAlarmModuleBuilder().DoDashboardBuilder().ToSelects(list),
		Pagination: builder.NewParamsBuild().PaginationModuleBuilder().ToAPI(params.Page),
	}, nil
}
