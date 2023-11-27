package alarmservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/alarm/page"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

type AlarmPageService struct {
	pb.UnimplementedAlarmPageServer

	log *log.Helper

	pageBiz *biz.PageBiz
}

// NewAlarmPageService 实例化AlarmPageService
func NewAlarmPageService(pageBiz *biz.PageBiz, logger log.Logger) *AlarmPageService {
	return &AlarmPageService{
		log:     log.NewHelper(log.With(logger, "module", "service.alarm.page")),
		pageBiz: pageBiz,
	}
}

func (s *AlarmPageService) CreateAlarmPage(ctx context.Context, req *pb.CreateAlarmPageRequest) (*pb.CreateAlarmPageReply, error) {
	alarmPageBO := &dobo.AlarmPageBO{
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
	alarmPageBO := &dobo.AlarmPageBO{
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

	alarmPageInfo := alarmPageBOToAlarmPageInfo(alarmPageBO)

	return &pb.GetAlarmPageReply{AlarmPage: alarmPageInfo}, nil
}

// alarmPageBOToAlarmPageInfo .
func alarmPageBOToAlarmPageInfo(alarmPageBO *dobo.AlarmPageBO) *api.AlarmPageV1 {
	if alarmPageBO == nil {
		return nil
	}
	return &api.AlarmPageV1{
		Id:     alarmPageBO.Id,
		Name:   alarmPageBO.Name,
		Icon:   alarmPageBO.Icon,
		Color:  alarmPageBO.Color,
		Status: api.Status(alarmPageBO.Status),
		Remark: alarmPageBO.Remark,
	}
}

// alarmPageBOToAlarmPageInfoSelect .
func alarmPageBOToAlarmPageInfoSelect(alarmPageBO *dobo.AlarmPageBO) *api.AlarmPageSelectV1 {
	if alarmPageBO == nil {
		return nil
	}
	return &api.AlarmPageSelectV1{
		Value:  alarmPageBO.Id,
		Label:  alarmPageBO.Name,
		Icon:   alarmPageBO.Icon,
		Color:  alarmPageBO.Color,
		Status: api.Status(alarmPageBO.Status),
		Remark: alarmPageBO.Remark,
	}
}

func (s *AlarmPageService) ListAlarmPage(ctx context.Context, req *pb.ListAlarmPageRequest) (*pb.ListAlarmPageReply, error) {
	alarmPageBOs, pgInfo, err := s.pageBiz.ListPage(ctx, req)
	if err != nil {
		return nil, err
	}

	list := make([]*api.AlarmPageV1, 0, len(alarmPageBOs))
	for _, alarmPageBO := range alarmPageBOs {
		alarmPageInfo := alarmPageBOToAlarmPageInfo(alarmPageBO)
		list = append(list, alarmPageInfo)
	}

	return &pb.ListAlarmPageReply{
		List: list,
		Page: &api.PageReply{
			Curr:  int32(pgInfo.GetCurr()),
			Size:  int32(pgInfo.GetSize()),
			Total: pgInfo.GetTotal(),
		},
	}, nil
}

func (s *AlarmPageService) SelectAlarmPage(ctx context.Context, req *pb.SelectAlarmPageRequest) (*pb.SelectAlarmPageReply, error) {
	alarmPageBOs, pgInfo, err := s.pageBiz.SelectPageList(ctx, req)
	if err != nil {
		return nil, err
	}

	list := make([]*api.AlarmPageSelectV1, 0, len(alarmPageBOs))
	for _, alarmPageBO := range alarmPageBOs {
		alarmPageInfo := alarmPageBOToAlarmPageInfoSelect(alarmPageBO)
		list = append(list, alarmPageInfo)
	}
	return &pb.SelectAlarmPageReply{
		List: list,
		Page: &api.PageReply{
			Curr:  int32(pgInfo.GetCurr()),
			Size:  int32(pgInfo.GetSize()),
			Total: pgInfo.GetTotal(),
		},
	}, nil
}
