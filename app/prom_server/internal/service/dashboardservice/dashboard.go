package dashboardservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/server/dashboard"
	"github.com/aide-family/moon/app/prom_server/internal/biz"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/helper/middler"
	"github.com/aide-family/moon/pkg/util/slices"
)

type DashboardService struct {
	dashboard.UnimplementedDashboardServer

	log          *log.Helper
	dashboardBiz *biz.DashboardBiz
}

func NewDashboardService(dashboardBiz *biz.DashboardBiz, logger log.Logger) *DashboardService {
	return &DashboardService{
		log:          log.NewHelper(log.With(logger, "module", "service.dashboard-service")),
		dashboardBiz: dashboardBiz,
	}
}

func (s *DashboardService) CreateDashboard(ctx context.Context, req *dashboard.CreateDashboardRequest) (*dashboard.CreateDashboardReply, error) {
	userId := middler.GetUserId(ctx)
	newDashboardBo := &bo.MyDashboardConfigBO{
		Remark: req.GetRemark(),
		Title:  req.GetTitle(),
		Color:  req.GetColor(),
		UserId: userId,
		Charts: slices.To(req.GetChartIds(), func(id uint32) *bo.MyChartBO {
			return &bo.MyChartBO{Id: id}
		}),
	}
	newDashboardBo, err := s.dashboardBiz.CreateDashboard(ctx, newDashboardBo)
	if err != nil {
		return nil, err
	}
	return &dashboard.CreateDashboardReply{
		Id: newDashboardBo.Id,
	}, nil
}

func (s *DashboardService) UpdateDashboard(ctx context.Context, req *dashboard.UpdateDashboardRequest) (*dashboard.UpdateDashboardReply, error) {
	newDashboard := &bo.MyDashboardConfigBO{
		Id:     req.GetId(),
		Remark: req.GetRemark(),
		Title:  req.GetTitle(),
		Color:  req.GetColor(),
		Charts: slices.To(req.GetChartIds(), func(id uint32) *bo.MyChartBO { return &bo.MyChartBO{Id: id} }),
	}
	newDashboardBo, err := s.dashboardBiz.UpdateDashboardById(ctx, newDashboard.Id, newDashboard)
	if err != nil {
		return nil, err
	}
	return &dashboard.UpdateDashboardReply{
		Id: newDashboardBo.Id,
	}, nil
}

func (s *DashboardService) DeleteDashboard(ctx context.Context, req *dashboard.DeleteDashboardRequest) (*dashboard.DeleteDashboardReply, error) {
	if err := s.dashboardBiz.DeleteDashboardById(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &dashboard.DeleteDashboardReply{
		Id: req.GetId(),
	}, nil
}

func (s *DashboardService) GetDashboard(ctx context.Context, req *dashboard.GetDashboardRequest) (*dashboard.GetDashboardReply, error) {
	detail, err := s.dashboardBiz.GetDashboardById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &dashboard.GetDashboardReply{
		Detail: detail.ToApi(),
	}, nil
}

func (s *DashboardService) ListDashboard(ctx context.Context, req *dashboard.ListDashboardRequest) (*dashboard.ListDashboardReply, error) {
	pgReq := req.GetPage()
	pgInfo := bo.NewPage(pgReq.GetCurr(), pgReq.GetSize())

	dashboardBoList, err := s.dashboardBiz.ListDashboard(ctx, &bo.ListDashboardReq{
		Page:    pgInfo,
		Keyword: req.GetKeyword(),
		Status:  vobj.Status(req.GetStatus()),
	})
	if err != nil {
		return nil, err
	}

	return &dashboard.ListDashboardReply{
		Page: &api.PageReply{
			Total: pgInfo.GetTotal(),
			Size:  pgReq.GetSize(),
			Curr:  pgReq.GetCurr(),
		},
		List: slices.To(dashboardBoList, func(i *bo.MyDashboardConfigBO) *api.MyDashboardConfig {
			return i.ToApi()
		}),
	}, nil
}

func (s *DashboardService) ListDashboardSelect(ctx context.Context, req *dashboard.ListDashboardSelectRequest) (*dashboard.ListDashboardSelectReply, error) {
	pgReq := req.GetPage()
	pgInfo := bo.NewPage(pgReq.GetCurr(), pgReq.GetSize())

	dashboardBoList, err := s.dashboardBiz.ListDashboard(ctx, &bo.ListDashboardReq{
		Page:    pgInfo,
		Keyword: req.GetKeyword(),
		Status:  vobj.Status(req.GetStatus()),
	})
	if err != nil {
		return nil, err
	}

	return &dashboard.ListDashboardSelectReply{
		Page: &api.PageReply{
			Total: pgInfo.GetTotal(),
			Size:  pgReq.GetSize(),
			Curr:  pgReq.GetCurr(),
		},
		List: slices.To(dashboardBoList, func(i *bo.MyDashboardConfigBO) *api.MyDashboardConfigOption {
			return i.ToApiSelectV1()
		}),
	}, nil
}
