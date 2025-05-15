package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/cmd/palace/internal/service/build"
	"github.com/moon-monitor/moon/pkg/api/palace"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
)

type TeamDashboardService struct {
	palace.UnimplementedTeamDashboardServer

	dashboard *biz.DashboardBiz
	helper    *log.Helper
}

func NewTeamDashboardService(dashboard *biz.DashboardBiz, logger log.Logger) *TeamDashboardService {
	return &TeamDashboardService{
		dashboard: dashboard,
		helper:    log.NewHelper(log.With(logger, "module", "service.teamDashboard")),
	}
}

func (s *TeamDashboardService) SaveTeamDashboard(ctx context.Context, req *palace.SaveTeamDashboardRequest) (*common.EmptyReply, error) {
	params := &bo.SaveDashboardReq{
		ID:       req.GetDashboardId(),
		Title:    req.GetTitle(),
		Remark:   req.GetRemark(),
		Status:   vobj.GlobalStatus(req.GetStatus()),
		ColorHex: req.GetColorHex(),
	}
	err := s.dashboard.SaveDashboard(ctx, params)
	if err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "保存团队图表看板成功"}, nil
}

func (s *TeamDashboardService) DeleteTeamDashboard(ctx context.Context, req *palace.DeleteTeamDashboardRequest) (*common.EmptyReply, error) {
	if err := s.dashboard.DeleteDashboard(ctx, req.GetDashboardId()); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "删除团队图表看板成功"}, nil
}

func (s *TeamDashboardService) GetTeamDashboard(ctx context.Context, req *palace.GetTeamDashboardRequest) (*common.TeamDashboardItem, error) {
	dashboard, err := s.dashboard.GetDashboard(ctx, req.GetDashboardId())
	if err != nil {
		return nil, err
	}
	return build.ToDashboardItem(dashboard), nil
}

func (s *TeamDashboardService) ListTeamDashboard(ctx context.Context, req *palace.ListTeamDashboardRequest) (*palace.ListTeamDashboardReply, error) {
	params := &bo.ListDashboardReq{
		PaginationRequest: build.ToPaginationRequest(req.Pagination),
		Status:            vobj.GlobalStatus(req.GetStatus()),
		Keyword:           req.GetKeyword(),
	}
	reply, err := s.dashboard.ListDashboard(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.ListTeamDashboardReply{
		Items:      build.ToDashboardItems(reply.Items),
		Pagination: build.ToPaginationReply(reply.PaginationReply),
	}, nil
}

func (s *TeamDashboardService) UpdateTeamDashboardStatus(ctx context.Context, req *palace.UpdateTeamDashboardStatusRequest) (*common.EmptyReply, error) {
	params := &bo.BatchUpdateDashboardStatusReq{
		Ids:    req.GetDashboardIds(),
		Status: vobj.GlobalStatus(req.GetStatus()),
	}
	err := s.dashboard.BatchUpdateDashboardStatus(ctx, params)
	if err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "更新团队图表看板状态成功"}, nil
}

func (s *TeamDashboardService) SaveTeamDashboardChart(ctx context.Context, req *palace.SaveTeamDashboardChartRequest) (*common.EmptyReply, error) {
	params := &bo.SaveDashboardChartReq{
		ID:          req.GetChartId(),
		DashboardID: req.GetDashboardId(),
		Title:       req.GetTitle(),
		Remark:      req.GetRemark(),
		Status:      vobj.GlobalStatus(req.GetStatus()),
		Url:         req.GetUrl(),
		Width:       req.GetWidth(),
		Height:      req.GetHeight(),
	}
	if err := s.dashboard.SaveDashboardChart(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "保存团队图表成功"}, nil
}

func (s *TeamDashboardService) DeleteTeamDashboardChart(ctx context.Context, req *palace.DeleteTeamDashboardChartRequest) (*common.EmptyReply, error) {
	params := &bo.OperateOneDashboardChartReq{
		ID:          req.GetChartId(),
		DashboardID: req.GetDashboardId(),
	}
	if err := s.dashboard.DeleteDashboardChart(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "删除团队图表成功"}, nil
}

func (s *TeamDashboardService) GetTeamDashboardChart(ctx context.Context, req *palace.GetTeamDashboardChartRequest) (*common.TeamDashboardChartItem, error) {
	params := &bo.OperateOneDashboardChartReq{
		ID:          req.GetChartId(),
		DashboardID: req.GetDashboardId(),
	}
	chart, err := s.dashboard.GetDashboardChart(ctx, params)
	if err != nil {
		return nil, err
	}
	return build.ToDashboardChartItem(chart), nil
}

func (s *TeamDashboardService) ListTeamDashboardChart(ctx context.Context, req *palace.ListTeamDashboardChartRequest) (*palace.ListTeamDashboardChartReply, error) {
	params := &bo.ListDashboardChartReq{
		PaginationRequest: build.ToPaginationRequest(req.Pagination),
		Status:            vobj.GlobalStatus(req.GetStatus()),
		DashboardID:       req.GetDashboardId(),
		Keyword:           req.GetKeyword(),
	}
	reply, err := s.dashboard.ListDashboardCharts(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.ListTeamDashboardChartReply{
		Items:      build.ToDashboardChartItems(reply.Items),
		Pagination: build.ToPaginationReply(reply.PaginationReply),
	}, nil
}

func (s *TeamDashboardService) UpdateTeamDashboardChartStatus(ctx context.Context, req *palace.UpdateTeamDashboardChartStatusRequest) (*common.EmptyReply, error) {
	params := &bo.BatchUpdateDashboardChartStatusReq{
		DashboardID: req.GetDashboardId(),
		Ids:         req.GetChartIds(),
		Status:      vobj.GlobalStatus(req.GetStatus()),
	}

	if err := s.dashboard.BatchUpdateDashboardChartStatus(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "更新团队图表状态成功"}, nil
}
