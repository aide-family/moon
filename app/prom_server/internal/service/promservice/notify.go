package promservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/vobj"

	"prometheus-manager/api"
	pb "prometheus-manager/api/server/prom/notify"
	"prometheus-manager/pkg/util/slices"

	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
)

type NotifyService struct {
	pb.UnimplementedNotifyServer

	log *log.Helper

	notifyBiz *biz.NotifyBiz
}

func NewNotifyService(notifyBiz *biz.NotifyBiz, logger log.Logger) *NotifyService {
	return &NotifyService{
		log:       log.NewHelper(log.With(logger, "module", "service.prom.notify")),
		notifyBiz: notifyBiz,
	}
}

func (s *NotifyService) CreateNotify(ctx context.Context, req *pb.CreateNotifyRequest) (*pb.CreateNotifyReply, error) {
	notifyBo := &bo.NotifyBO{
		Name:   req.GetName(),
		Remark: req.GetRemark(),
		BeNotifyMembers: slices.To(req.GetMembers(), func(t *api.BeNotifyMember) *bo.NotifyMemberBO {
			return bo.NotifyMemberApiToBO(t)
		}),
		ChatGroups: slices.To(req.GetChatGroups(), func(t uint32) *bo.ChatGroupBO {
			chatGroupApi := &api.ChatGroup{
				Id: t,
			}
			return bo.ChatGroupApiToBO(chatGroupApi)
		}),
	}

	// 判断通知对象名称是否已存在
	if err := s.notifyBiz.CheckNotifyName(ctx, notifyBo.Name); err != nil {
		return nil, err
	}

	// TODO 检查member是否存在

	notifyBo, err := s.notifyBiz.CreateNotify(ctx, notifyBo)
	if err != nil {
		return nil, err
	}

	return &pb.CreateNotifyReply{
		Id: notifyBo.Id,
	}, nil
}

func (s *NotifyService) UpdateNotify(ctx context.Context, req *pb.UpdateNotifyRequest) (*pb.UpdateNotifyReply, error) {
	notifyBo := &bo.NotifyBO{
		Id:     req.GetId(),
		Name:   req.GetName(),
		Remark: req.GetRemark(),
		Status: vobj.Status(req.GetStatus()),
		BeNotifyMembers: slices.To(req.GetMembers(), func(t *api.BeNotifyMember) *bo.NotifyMemberBO {
			return bo.NotifyMemberApiToBO(t)
		}),
		ChatGroups: slices.To(req.GetChatGroups(), func(t uint32) *bo.ChatGroupBO {
			chatGroupApi := &api.ChatGroup{
				Id: t,
			}
			return bo.ChatGroupApiToBO(chatGroupApi)
		}),
	}

	if err := s.notifyBiz.CheckNotifyName(ctx, notifyBo.Name, notifyBo.Id); err != nil {
		return nil, err
	}

	// TODO 检查member是否存在

	if err := s.notifyBiz.UpdateNotifyById(ctx, notifyBo.Id, notifyBo); err != nil {
		return nil, err
	}
	return &pb.UpdateNotifyReply{
		Id: req.GetId(),
	}, nil
}

func (s *NotifyService) DeleteNotify(ctx context.Context, req *pb.DeleteNotifyRequest) (*pb.DeleteNotifyReply, error) {
	if err := s.notifyBiz.DeleteNotifyById(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &pb.DeleteNotifyReply{Id: req.GetId()}, nil
}

func (s *NotifyService) GetNotify(ctx context.Context, req *pb.GetNotifyRequest) (*pb.GetNotifyReply, error) {
	notifyBo, err := s.notifyBiz.GetNotifyById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.GetNotifyReply{
		Detail: notifyBo.ToApi(),
	}, nil
}

func (s *NotifyService) ListNotify(ctx context.Context, req *pb.ListNotifyRequest) (*pb.ListNotifyReply, error) {
	pgReq := req.GetPage()
	pgInfo := bo.NewPage(pgReq.GetCurr(), pgReq.GetSize())

	notifyBos, err := s.notifyBiz.ListNotify(ctx, &bo.ListNotifyRequest{
		Page:    pgInfo,
		Keyword: req.GetKeyword(),
		Status:  vobj.Status(req.GetStatus()),
	})
	if err != nil {
		return nil, err
	}

	list := slices.To(notifyBos, func(t *bo.NotifyBO) *api.NotifyV1 {
		return t.ToApi()
	})
	return &pb.ListNotifyReply{
		Page: &api.PageReply{
			Curr:  pgInfo.GetRespCurr(),
			Size:  pgInfo.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}

func (s *NotifyService) SelectNotify(ctx context.Context, req *pb.SelectNotifyRequest) (*pb.SelectNotifyReply, error) {
	pgReq := req.GetPage()
	pgInfo := bo.NewPage(pgReq.GetCurr(), pgReq.GetSize())

	notifyBos, err := s.notifyBiz.ListNotify(ctx, &bo.ListNotifyRequest{
		Page:    pgInfo,
		Keyword: req.GetKeyword(),
		Status:  vobj.Status(req.GetStatus()),
	})
	if err != nil {
		return nil, err
	}

	list := slices.To(notifyBos, func(t *bo.NotifyBO) *api.NotifySelectV1 {
		return t.ToApiSelectV1()
	})
	return &pb.SelectNotifyReply{
		Page: &api.PageReply{
			Curr:  pgInfo.GetRespCurr(),
			Size:  pgInfo.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}
