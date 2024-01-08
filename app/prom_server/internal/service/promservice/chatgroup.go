package promservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/api"
	pb "prometheus-manager/api/prom/notify"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

type ChatGroupService struct {
	pb.UnimplementedChatGroupServer

	log          *log.Helper
	chatGroupBiz *biz.ChatGroupBiz
}

func NewChatGroupService(chatGroupBiz *biz.ChatGroupBiz, logger log.Logger) *ChatGroupService {
	return &ChatGroupService{
		log:          log.NewHelper(log.With(logger, "module", "service.prom.chatgroup")),
		chatGroupBiz: chatGroupBiz,
	}
}

func (s *ChatGroupService) CreateChatGroup(ctx context.Context, req *pb.CreateChatGroupRequest) (*pb.CreateChatGroupReply, error) {
	chatGroupBo := &bo.ChatGroupBO{
		Name:      req.GetName(),
		Remark:    req.GetRemark(),
		Hook:      req.GetHook(),
		NotifyApp: vo.NotifyApp(req.GetApp()),
		HookName:  req.GetHookName(),
	}
	chatGroupBo, err := s.chatGroupBiz.CreateChatGroup(ctx, chatGroupBo)
	if err != nil {
		s.log.Errorf("CreateChatGroup err: %v", err)
		return nil, err
	}
	return &pb.CreateChatGroupReply{
		Id: chatGroupBo.Id,
	}, nil
}

func (s *ChatGroupService) UpdateChatGroup(ctx context.Context, req *pb.UpdateChatGroupRequest) (*pb.UpdateChatGroupReply, error) {
	chatGroupBo := &bo.ChatGroupBO{
		Id:        req.GetId(),
		Name:      req.GetName(),
		Remark:    req.GetRemark(),
		Status:    vo.Status(req.GetStatus()),
		Hook:      req.GetHook(),
		NotifyApp: vo.NotifyApp(req.GetApp()),
		HookName:  req.GetHookName(),
	}
	if err := s.chatGroupBiz.UpdateChatGroupById(ctx, chatGroupBo, chatGroupBo.Id); err != nil {
		s.log.Errorf("UpdateChatGroup err: %v", err)
		return nil, err
	}
	return &pb.UpdateChatGroupReply{Id: chatGroupBo.Id}, nil
}

func (s *ChatGroupService) DeleteChatGroup(ctx context.Context, req *pb.DeleteChatGroupRequest) (*pb.DeleteChatGroupReply, error) {
	if err := s.chatGroupBiz.DeleteChatGroupById(ctx, req.GetId()); err != nil {
		s.log.Errorf("DeleteChatGroup err: %v", err)
		return nil, err
	}
	return &pb.DeleteChatGroupReply{}, nil
}

func (s *ChatGroupService) GetChatGroup(ctx context.Context, req *pb.GetChatGroupRequest) (*pb.GetChatGroupReply, error) {
	chatGroupBo, err := s.chatGroupBiz.GetChatGroupById(ctx, req.GetId())
	if err != nil {
		s.log.Errorf("GetChatGroup err: %v", err)
		return nil, err
	}
	return &pb.GetChatGroupReply{
		Detail: chatGroupBo.ToApi(),
	}, nil
}

func (s *ChatGroupService) ListChatGroup(ctx context.Context, req *pb.ListChatGroupRequest) (*pb.ListChatGroupReply, error) {
	pgReq := req.GetPage()
	pgInfo := basescopes.NewPage(pgReq.GetCurr(), pgReq.GetSize())

	wheres := []basescopes.ScopeMethod{
		basescopes.NameLike(req.GetKeyword()),
		basescopes.StatusEQ(vo.Status(req.GetStatus())),
	}

	chatGroupBos, err := s.chatGroupBiz.ListChatGroup(ctx, pgInfo, wheres...)
	if err != nil {
		s.log.Errorf("ListChatGroup err: %v", err)
		return nil, err
	}
	return &pb.ListChatGroupReply{
		Page: &api.PageReply{
			Curr:  pgInfo.GetCurr(),
			Size:  pgInfo.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: slices.To(chatGroupBos, func(i *bo.ChatGroupBO) *api.ChatGroup {
			return i.ToApi()
		}),
	}, nil
}

func (s *ChatGroupService) SelectChatGroup(ctx context.Context, req *pb.SelectChatGroupRequest) (*pb.SelectChatGroupReply, error) {
	pgReq := req.GetPage()
	pgInfo := basescopes.NewPage(pgReq.GetCurr(), pgReq.GetSize())

	wheres := []basescopes.ScopeMethod{
		basescopes.NameLike(req.GetKeyword()),
		basescopes.StatusEQ(vo.Status(req.GetStatus())),
	}

	chatGroupBos, err := s.chatGroupBiz.ListChatGroup(ctx, pgInfo, wheres...)
	if err != nil {
		s.log.Errorf("ListChatGroup err: %v", err)
		return nil, err
	}
	return &pb.SelectChatGroupReply{
		Page: &api.PageReply{
			Curr:  pgInfo.GetCurr(),
			Size:  pgInfo.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: slices.To(chatGroupBos, func(i *bo.ChatGroupBO) *api.ChatGroupSelectV1 {
			return i.ToSelectApi()
		}),
	}, nil
}
