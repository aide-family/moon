package promservice

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/prom/notify"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/pkg/helper/model/notifyscopes"
	"prometheus-manager/pkg/util/slices"
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
	notifyBo := &dobo.NotifyBO{
		Name:   req.GetName(),
		Remark: req.GetRemark(),
		BeNotifyMembers: slices.To(req.GetMembers(), func(t *api.BeNotifyMember) *dobo.NotifyMemberBO {
			return dobo.NotifyMemberApiToBO(t)
		}),
		ChatGroups: slices.To(req.GetChatGroups(), func(t uint32) *dobo.ChatGroupBO {
			chatGroupApi := &api.ChatGroup{
				Id: t,
			}
			return dobo.ChatGroupApiToBO(chatGroupApi)
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
		Id: uint32(notifyBo.Id),
	}, nil
}

func (s *NotifyService) UpdateNotify(ctx context.Context, req *pb.UpdateNotifyRequest) (*pb.UpdateNotifyReply, error) {
	notifyBo := &dobo.NotifyBO{
		Id:     uint(req.GetId()),
		Name:   req.GetName(),
		Remark: req.GetRemark(),
		Status: int32(req.GetStatus()),
		BeNotifyMembers: slices.To(req.GetMembers(), func(t *api.BeNotifyMember) *dobo.NotifyMemberBO {
			return dobo.NotifyMemberApiToBO(t)
		}),
		ChatGroups: slices.To(req.GetChatGroups(), func(t uint32) *dobo.ChatGroupBO {
			chatGroupApi := &api.ChatGroup{
				Id: t,
			}
			return dobo.ChatGroupApiToBO(chatGroupApi)
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
	if err := s.notifyBiz.DeleteNotifyById(ctx, uint(req.GetId())); err != nil {
		return nil, err
	}
	return &pb.DeleteNotifyReply{Id: req.GetId()}, nil
}

func (s *NotifyService) GetNotify(ctx context.Context, req *pb.GetNotifyRequest) (*pb.GetNotifyReply, error) {
	notifyBo, err := s.notifyBiz.GetNotifyById(ctx, uint(req.GetId()))
	if err != nil {
		return nil, err
	}
	return &pb.GetNotifyReply{
		Detail: notifyBo.ToApi(),
		Response: &api.Response{
			Code: 0,
			Msg:  "",
		},
	}, nil
}

func (s *NotifyService) ListNotify(ctx context.Context, req *pb.ListNotifyRequest) (*pb.ListNotifyReply, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	wheres := []query.ScopeMethod{
		notifyscopes.NotifyLike(req.GetKeyword()),
	}
	notifyBos, err := s.notifyBiz.ListNotify(ctx, pgInfo, wheres...)
	if err != nil {
		return nil, err
	}

	list := slices.To(notifyBos, func(t *dobo.NotifyBO) *api.NotifyV1 {
		return t.ToApi()
	})
	return &pb.ListNotifyReply{
		Page: &api.PageReply{
			Curr:  pgReq.GetCurr(),
			Size:  pgReq.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
		Response: &api.Response{
			Code: 0,
			Msg:  "",
		},
	}, nil
}

func (s *NotifyService) SelectNotify(ctx context.Context, req *pb.SelectNotifyRequest) (*pb.SelectNotifyReply, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	wheres := []query.ScopeMethod{
		notifyscopes.NotifyLike(req.GetKeyword()),
	}
	notifyBos, err := s.notifyBiz.ListNotify(ctx, pgInfo, wheres...)
	if err != nil {
		return nil, err
	}

	list := slices.To(notifyBos, func(t *dobo.NotifyBO) *api.NotifySelectV1 {
		return t.ToApiSelectV1()
	})
	return &pb.SelectNotifyReply{
		Page: &api.PageReply{
			Curr:  pgReq.GetCurr(),
			Size:  pgReq.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
		Response: &api.Response{
			Code: 0,
			Msg:  "",
		},
	}, nil
}
