package alarmservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/helper/middler"
	"prometheus-manager/pkg/util/slices"

	"prometheus-manager/api"
	pb "prometheus-manager/api/alarm/page"

	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
)

type AlarmPageService struct {
	pb.UnimplementedAlarmPageServer

	log *log.Helper

	pageBiz *biz.AlarmPageBiz
}

// NewAlarmPageService 实例化AlarmPageService
func NewAlarmPageService(pageBiz *biz.AlarmPageBiz, logger log.Logger) *AlarmPageService {
	return &AlarmPageService{
		log:     log.NewHelper(log.With(logger, "module", "service.alarm.page")),
		pageBiz: pageBiz,
	}
}

func (s *AlarmPageService) CreateAlarmPage(ctx context.Context, req *pb.CreateAlarmPageRequest) (*pb.CreateAlarmPageReply, error) {
	alarmPageBO := &bo.AlarmPageBO{
		Name:   req.GetName(),
		Remark: req.GetRemark(),
		Icon:   req.GetIcon(),
		Color:  req.GetColor(),
	}

	alarmPageBO, err := s.pageBiz.CreatePage(ctx, alarmPageBO)
	if err != nil {
		return nil, err
	}

	return &pb.CreateAlarmPageReply{Id: alarmPageBO.Id}, nil
}

func (s *AlarmPageService) UpdateAlarmPage(ctx context.Context, req *pb.UpdateAlarmPageRequest) (*pb.UpdateAlarmPageReply, error) {
	alarmPageBO := &bo.AlarmPageBO{
		Id:     req.GetId(),
		Name:   req.GetName(),
		Remark: req.GetRemark(),
		Icon:   req.GetIcon(),
		Color:  req.GetColor(),
	}

	alarmPageBO, err := s.pageBiz.UpdatePage(ctx, alarmPageBO)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateAlarmPageReply{Id: alarmPageBO.Id}, nil
}

func (s *AlarmPageService) BatchUpdateAlarmPageStatus(ctx context.Context, req *pb.BatchUpdateAlarmPageStatusRequest) (*pb.BatchUpdateAlarmPageStatusReply, error) {
	updateAlarmPageIds := req.GetIds()
	if err := s.pageBiz.BatchUpdatePageStatusByIds(ctx, req.GetStatus(), updateAlarmPageIds); err != nil {
		return nil, err
	}
	return &pb.BatchUpdateAlarmPageStatusReply{Ids: updateAlarmPageIds}, nil
}

func (s *AlarmPageService) DeleteAlarmPage(ctx context.Context, req *pb.DeleteAlarmPageRequest) (*pb.DeleteAlarmPageReply, error) {
	if err := s.pageBiz.DeletePageByIds(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &pb.DeleteAlarmPageReply{Id: req.GetId()}, nil
}

func (s *AlarmPageService) BatchDeleteAlarmPage(ctx context.Context, req *pb.BatchDeleteAlarmPageRequest) (*pb.BatchDeleteAlarmPageReply, error) {
	if err := s.pageBiz.DeletePageByIds(ctx, req.GetIds()...); err != nil {
		return nil, err
	}
	return &pb.BatchDeleteAlarmPageReply{Ids: req.GetIds()}, nil
}

func (s *AlarmPageService) GetAlarmPage(ctx context.Context, req *pb.GetAlarmPageRequest) (*pb.GetAlarmPageReply, error) {
	alarmPageBO, err := s.pageBiz.GetPageById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.GetAlarmPageReply{AlarmPage: alarmPageBO.ToApi()}, nil
}

func (s *AlarmPageService) ListAlarmPage(ctx context.Context, req *pb.ListAlarmPageRequest) (*pb.ListAlarmPageReply, error) {
	alarmPageBOs, pgInfo, err := s.pageBiz.ListPage(ctx, req)
	if err != nil {
		return nil, err
	}

	list := slices.To(alarmPageBOs, func(alarmPageBO *bo.AlarmPageBO) *api.AlarmPageV1 {
		return alarmPageBO.ToApi()
	})

	return &pb.ListAlarmPageReply{
		List: list,
		Page: &api.PageReply{
			Curr:  pgInfo.GetCurr(),
			Size:  pgInfo.GetSize(),
			Total: pgInfo.GetTotal(),
		},
	}, nil
}

func (s *AlarmPageService) SelectAlarmPage(ctx context.Context, req *pb.SelectAlarmPageRequest) (*pb.SelectAlarmPageReply, error) {
	alarmPageBOs, pgInfo, err := s.pageBiz.SelectPageList(ctx, req)
	if err != nil {
		return nil, err
	}

	list := slices.To(alarmPageBOs, func(alarmPageBO *bo.AlarmPageBO) *api.AlarmPageSelectV1 {
		return alarmPageBO.ToApiSelectV1()
	})
	return &pb.SelectAlarmPageReply{
		List: list,
		Page: &api.PageReply{
			Curr:  pgInfo.GetCurr(),
			Size:  pgInfo.GetSize(),
			Total: pgInfo.GetTotal(),
		},
	}, nil
}

// CountAlarmPage 统计各告警页面的告警数量
func (s *AlarmPageService) CountAlarmPage(ctx context.Context, req *pb.CountAlarmPageRequest) (*pb.CountAlarmPageReply, error) {
	count, err := s.pageBiz.CountAlarmPageByIds(ctx, req.GetIds()...)
	if err != nil {
		return nil, err
	}
	return &pb.CountAlarmPageReply{
		AlarmCount: count,
	}, nil
}

// ListMyAlarmPage 获取我的告警页面列表
func (s *AlarmPageService) ListMyAlarmPage(ctx context.Context, _ *pb.ListMyAlarmPageRequest) (*pb.ListMyAlarmPageReply, error) {
	userId := middler.GetUserId(ctx)
	userAlarmPages, err := s.pageBiz.GetUserAlarmPages(ctx, userId)
	if err != nil {
		return nil, err
	}
	return &pb.ListMyAlarmPageReply{
		List: slices.To(userAlarmPages, func(alarmPageBO *bo.AlarmPageBO) *api.AlarmPageV1 {
			return alarmPageBO.ToApi()
		}),
	}, nil
}

// MyAlarmPagesConfig 我的告警页面列表配置
func (s *AlarmPageService) MyAlarmPagesConfig(ctx context.Context, req *pb.MyAlarmPagesConfigRequest) (*pb.MyAlarmPagesConfigReply, error) {
	userId := middler.GetUserId(ctx)
	if err := s.pageBiz.BindUserPages(ctx, userId, req.GetAlarmIds()); err != nil {
		return nil, err
	}
	return &pb.MyAlarmPagesConfigReply{}, nil
}
