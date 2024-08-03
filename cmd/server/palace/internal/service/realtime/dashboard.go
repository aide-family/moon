package realtime

import (
	"context"

	realtimeapi "github.com/aide-family/moon/api/admin/realtime"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
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
	params := build.NewBuilder().WithContext(ctx).DashboardModule().WithAPIAddDashboardParams(req).ToBo()
	if err := s.dashboardBiz.CreateDashboard(ctx, params); err != nil {
		return nil, err
	}
	return &realtimeapi.CreateDashboardReply{}, nil
}

// UpdateDashboard 更新监控大盘
func (s *DashboardService) UpdateDashboard(ctx context.Context, req *realtimeapi.UpdateDashboardRequest) (*realtimeapi.UpdateDashboardReply, error) {
	params := build.NewBuilder().
		WithContext(ctx).
		DashboardModule().
		WithAPIUpdateDashboardParams(req).
		ToBo()
	if err := s.dashboardBiz.UpdateDashboard(ctx, params); err != nil {
		return nil, err
	}
	return &realtimeapi.UpdateDashboardReply{}, nil
}

// DeleteDashboard 删除监控大盘
func (s *DashboardService) DeleteDashboard(ctx context.Context, req *realtimeapi.DeleteDashboardRequest) (*realtimeapi.DeleteDashboardReply, error) {
	params := build.NewBuilder().
		WithContext(ctx).
		DashboardModule().
		WithAPIDeleteDashboardParams(req).
		ToBo()
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
	apiDetail := build.NewBuilder().
		WithContext(ctx).
		DashboardModule().
		WithDoDashboard(detail).
		ToAPI()
	return &realtimeapi.GetDashboardReply{
		Detail: apiDetail,
	}, nil
}

// ListDashboard 获取监控大盘列表
func (s *DashboardService) ListDashboard(ctx context.Context, req *realtimeapi.ListDashboardRequest) (*realtimeapi.ListDashboardReply, error) {
	params := build.NewBuilder().
		WithContext(ctx).
		DashboardModule().
		WithAPIQueryDashboardListParams(req).
		ToBo()
	list, err := s.dashboardBiz.ListDashboard(ctx, params)
	if err != nil {
		return nil, err
	}
	apiList := build.NewBuilder().
		WithContext(ctx).
		DashboardModule().
		WithDoDashboardList(list).
		ToAPIs()
	return &realtimeapi.ListDashboardReply{
		List:       apiList,
		Pagination: build.NewPageBuilder(params.Page).ToAPI(),
	}, nil
}

// ListDashboardSelect 获取监控大盘下拉列表
func (s *DashboardService) ListDashboardSelect(ctx context.Context, req *realtimeapi.ListDashboardSelectRequest) (*realtimeapi.ListDashboardSelectReply, error) {
	params := build.NewBuilder().
		WithContext(ctx).
		DashboardModule().
		WithAPIQueryDashboardSelectParams(req).
		ToBo()
	list, err := s.dashboardBiz.ListDashboard(ctx, params)
	if err != nil {
		return nil, err
	}
	apiList := build.NewBuilder().
		WithContext(ctx).
		DashboardModule().
		WithDoDashboardList(list).
		ToSelects()
	return &realtimeapi.ListDashboardSelectReply{
		List:       apiList,
		Pagination: build.NewPageBuilder(params.Page).ToAPI(),
	}, nil
}
