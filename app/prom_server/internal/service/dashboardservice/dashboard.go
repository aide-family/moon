package dashboardservice

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/server/dashboard"
	"github.com/aide-family/moon/app/prom_server/internal/biz"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg"
	"github.com/aide-family/moon/pkg/helper/middler"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/go-kratos/kratos/v2/log"
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
	newDashboardBo := &bo.CreateMyDashboardBO{
		Status: vobj.StatusEnabled,
		Remark: req.GetRemark(),
		Title:  req.GetTitle(),
		Color:  req.GetColor(),
		UserId: userId,
		Charts: slices.To(req.GetChartIds(), func(id uint32) *bo.MyChartBO {
			return &bo.MyChartBO{Id: id}
		}),
	}
	newDashboardDo, err := s.dashboardBiz.CreateDashboard(ctx, newDashboardBo)
	if err != nil {
		return nil, err
	}
	return &dashboard.CreateDashboardReply{
		Id: newDashboardDo.ID,
	}, nil
}

func (s *DashboardService) UpdateDashboard(ctx context.Context, req *dashboard.UpdateDashboardRequest) (*dashboard.UpdateDashboardReply, error) {
	newDashboard := &bo.UpdateMyDashboardBO{
		Id: req.GetId(),
		CreateMyDashboardBO: bo.CreateMyDashboardBO{
			Status: vobj.StatusEnabled,
			Remark: req.GetRemark(),
			Title:  req.GetTitle(),
			Color:  req.GetColor(),
			Charts: slices.To(req.GetChartIds(), func(id uint32) *bo.MyChartBO {
				return &bo.MyChartBO{Id: id}
			}),
		},
	}
	newDashboardDo, err := s.dashboardBiz.UpdateDashboardById(ctx, newDashboard.Id, newDashboard)
	if err != nil {
		return nil, err
	}
	return &dashboard.UpdateDashboardReply{
		Id: newDashboardDo.ID,
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
		Detail: dashboardDoToApi(detail),
	}, nil
}

// dashboardDoToApi dashboardDoToApi
func dashboardDoToApi(dashboardDo *do.MyDashboardConfig) *api.MyDashboardConfig {
	if pkg.IsNil(dashboardDo) {
		return nil
	}
	return &api.MyDashboardConfig{
		Id:        dashboardDo.ID,
		Title:     dashboardDo.Title,
		Remark:    dashboardDo.Remark,
		CreatedAt: dashboardDo.CreatedAt.Unix(),
		UpdatedAt: dashboardDo.UpdatedAt.Unix(),
		DeletedAt: int64(dashboardDo.DeletedAt),
		Color:     dashboardDo.Color,
		Charts: slices.To(dashboardDo.GetCharts(), func(chart *do.MyChart) *api.MyChart {
			return chartDoToApi(chart)
		}),
		Status: int32(dashboardDo.Status),
	}
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
		List: slices.To(dashboardBoList, func(i *do.MyDashboardConfig) *api.MyDashboardConfig {
			return dashboardDoToApi(i)
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
		List: slices.To(dashboardBoList, func(i *do.MyDashboardConfig) *api.MyDashboardConfigOption {
			return &api.MyDashboardConfigOption{
				Value: i.ID,
				Label: i.Title,
				Color: i.Color,
			}
		}),
	}, nil
}
