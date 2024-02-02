package dashboardservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/dashboard"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/helper/middler"
	"prometheus-manager/pkg/util/slices"
)

type DashboardService struct {
	pb.UnimplementedDashboardServer

	log          *log.Helper
	dashboardBiz *biz.DashboardBiz
}

func NewDashboardService(dashboardBiz *biz.DashboardBiz, logger log.Logger) *DashboardService {
	return &DashboardService{
		log:          log.NewHelper(log.With(logger, "module", "service.dashboard-service")),
		dashboardBiz: dashboardBiz,
	}
}

func (s *DashboardService) CreateDashboard(ctx context.Context, req *pb.CreateDashboardRequest) (*pb.CreateDashboardReply, error) {
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
	return &pb.CreateDashboardReply{
		Id: newDashboardBo.Id,
	}, nil
}

func (s *DashboardService) UpdateDashboard(ctx context.Context, req *pb.UpdateDashboardRequest) (*pb.UpdateDashboardReply, error) {
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
	return &pb.UpdateDashboardReply{
		Id: newDashboardBo.Id,
	}, nil
}

func (s *DashboardService) DeleteDashboard(ctx context.Context, req *pb.DeleteDashboardRequest) (*pb.DeleteDashboardReply, error) {
	if err := s.dashboardBiz.DeleteDashboardById(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &pb.DeleteDashboardReply{
		Id: req.GetId(),
	}, nil
}

func (s *DashboardService) GetDashboard(ctx context.Context, req *pb.GetDashboardRequest) (*pb.GetDashboardReply, error) {
	detail, err := s.dashboardBiz.GetDashboardById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.GetDashboardReply{
		Detail: detail.ToApi(),
	}, nil
}

func (s *DashboardService) ListDashboard(ctx context.Context, req *pb.ListDashboardRequest) (*pb.ListDashboardReply, error) {
	pgReq := req.GetPage()
	pgInfo := basescopes.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	wheres := []basescopes.ScopeMethod{
		basescopes.TitleLike(req.GetKeyword()),
		basescopes.StatusEQ(vo.Status(req.GetStatus())),
		basescopes.CreatedAtDesc(),
		basescopes.DeletedAtDesc(),
	}
	dashboardBoList, err := s.dashboardBiz.ListDashboard(ctx, pgInfo, wheres...)
	if err != nil {
		return nil, err
	}

	return &pb.ListDashboardReply{
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

func (s *DashboardService) ListDashboardSelect(ctx context.Context, req *pb.ListDashboardSelectRequest) (*pb.ListDashboardSelectReply, error) {
	pgReq := req.GetPage()
	pgInfo := basescopes.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	wheres := []basescopes.ScopeMethod{
		basescopes.TitleLike(req.GetKeyword()),
		basescopes.StatusEQ(vo.Status(req.GetStatus())),
		basescopes.CreatedAtDesc(),
		basescopes.DeletedAtDesc(),
	}
	dashboardBoList, err := s.dashboardBiz.ListDashboard(ctx, pgInfo, wheres...)
	if err != nil {
		return nil, err
	}

	return &pb.ListDashboardSelectReply{
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
