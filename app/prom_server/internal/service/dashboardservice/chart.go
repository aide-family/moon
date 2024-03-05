package dashboardservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	"prometheus-manager/api/server/dashboard"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/helper/middler"
	"prometheus-manager/pkg/util/slices"
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
		UserId: userId,
		Title:  req.GetTitle(),
		Remark: req.GetRemark(),
		Url:    req.GetUrl(),
	}
	newChartBo, err := s.dashboardBiz.CreateChart(ctx, newChartBo)
	if err != nil {
		return nil, err
	}
	return &dashboard.CreateChartReply{
		Id: newChartBo.Id,
	}, nil
}

func (s *ChartService) UpdateChart(ctx context.Context, req *dashboard.UpdateChartRequest) (*dashboard.UpdateChartReply, error) {
	newChartBo := &bo.MyChartBO{
		Id:     req.GetId(),
		Title:  req.GetTitle(),
		Remark: req.GetRemark(),
		Url:    req.GetUrl(),
	}
	newChartBo, err := s.dashboardBiz.UpdateChartById(ctx, newChartBo.Id, newChartBo)
	if err != nil {
		return nil, err
	}
	return &dashboard.UpdateChartReply{
		Id: newChartBo.Id,
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
		Detail: chartDetail.ToApi(),
	}, nil
}

func (s *ChartService) ListChart(ctx context.Context, req *dashboard.ListChartRequest) (*dashboard.ListChartReply, error) {
	pgReq := req.GetPage()
	pgInfo := basescopes.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	wheres := []basescopes.ScopeMethod{
		basescopes.StatusEQ(vo.Status(req.GetStatus())),
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
		List: slices.To(chartList, func(i *bo.MyChartBO) *api.MyChart { return i.ToApi() }),
	}, nil
}
