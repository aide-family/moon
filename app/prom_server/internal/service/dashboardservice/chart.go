package dashboardservice

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/server/dashboard"
	"github.com/aide-family/moon/app/prom_server/internal/biz"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg"
	"github.com/aide-family/moon/pkg/helper/middler"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/go-kratos/kratos/v2/log"
)

type ChartService struct {
	dashboard.UnimplementedChartServer

	log          *log.Helper
	dashboardBiz *biz.DashboardBiz
}

func NewChartService(dashboardBiz *biz.DashboardBiz, logger log.Logger) *ChartService {
	return &ChartService{
		log:          log.NewHelper(log.With(logger, "module", "service.chart-service")),
		dashboardBiz: dashboardBiz,
	}
}

func (s *ChartService) CreateChart(ctx context.Context, req *dashboard.CreateChartRequest) (*dashboard.CreateChartReply, error) {
	userId := middler.GetUserId(ctx)
	newChartBo := &bo.MyChartBO{
		Id:        0,
		UserId:    userId,
		Title:     req.GetTitle(),
		Remark:    req.GetRemark(),
		Url:       req.GetUrl(),
		Status:    vobj.StatusEnabled,
		ChartType: vobj.ChartType(req.GetChartType()),
		Width:     req.GetWidth(),
		Height:    req.GetHeight(),
	}
	newChartDo, err := s.dashboardBiz.CreateChart(ctx, newChartBo)
	if err != nil {
		return nil, err
	}
	return &dashboard.CreateChartReply{
		Id: newChartDo.ID,
	}, nil
}

func (s *ChartService) UpdateChart(ctx context.Context, req *dashboard.UpdateChartRequest) (*dashboard.UpdateChartReply, error) {
	newChartBo := &bo.MyChartBO{
		Id:        req.GetId(),
		Title:     req.GetTitle(),
		Remark:    req.GetRemark(),
		Url:       req.GetUrl(),
		ChartType: vobj.ChartType(req.GetChartType()),
		Width:     req.GetWidth(),
		Height:    req.GetHeight(),
	}
	newChartDo, err := s.dashboardBiz.UpdateChartById(ctx, newChartBo.Id, newChartBo)
	if err != nil {
		return nil, err
	}
	return &dashboard.UpdateChartReply{
		Id: newChartDo.ID,
	}, nil
}

func (s *ChartService) DeleteChart(ctx context.Context, req *dashboard.DeleteChartRequest) (*dashboard.DeleteChartReply, error) {
	if err := s.dashboardBiz.DeleteChartById(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &dashboard.DeleteChartReply{
		Id: req.GetId(),
	}, nil
}

func (s *ChartService) GetChart(ctx context.Context, req *dashboard.GetChartRequest) (*dashboard.GetChartReply, error) {
	chartDetail, err := s.dashboardBiz.GetChartDetail(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &dashboard.GetChartReply{
		Detail: chartDoToApi(chartDetail),
	}, nil
}

// chartDoToApi 转换为 api.MyChart
func chartDoToApi(chartDo *do.MyChart) *api.MyChart {
	if pkg.IsNil(chartDo) {
		return nil
	}
	return &api.MyChart{
		Title:     chartDo.Title,
		Remark:    chartDo.Remark,
		Url:       chartDo.Url,
		Id:        chartDo.ID,
		Status:    chartDo.Status.Value(),
		ChartType: api.ChartType(chartDo.ChartType),
		Width:     chartDo.Width,
		Height:    chartDo.Height,
	}
}

func (s *ChartService) ListChart(ctx context.Context, req *dashboard.ListChartRequest) (*dashboard.ListChartReply, error) {
	pgReq := req.GetPage()
	pgInfo := bo.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	wheres := []basescopes.ScopeMethod{
		basescopes.StatusEQ(vobj.Status(req.GetStatus())),
		basescopes.TitleLike(req.GetKeyword()),
		basescopes.CreatedAtDesc(),
		basescopes.UpdateAtDesc(),
	}
	chartList, err := s.dashboardBiz.ListChartByPage(ctx, pgInfo, wheres...)
	if err != nil {
		return nil, err
	}
	return &dashboard.ListChartReply{
		Page: &api.PageReply{
			Curr:  pgReq.GetCurr(),
			Size:  pgReq.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: slices.To(chartList, func(i *do.MyChart) *api.MyChart { return chartDoToApi(i) }),
	}, nil
}
