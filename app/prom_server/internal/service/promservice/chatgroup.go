package promservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/aide-family/moon/api"
	pb "github.com/aide-family/moon/api/server/prom/notify"
	"github.com/aide-family/moon/app/prom_server/internal/biz"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
)

type ChatGroupService struct {
	pb.UnimplementedChatGroupServer

	log          *log.Helper
	chatGroupBiz *biz.ChatGroupBiz
	notifyBiz    *biz.NotifyBiz
}

func NewChatGroupService(
	chatGroupBiz *biz.ChatGroupBiz,
	notifyBiz *biz.NotifyBiz,
	logger log.Logger,
) *ChatGroupService {
	return &ChatGroupService{
		log:          log.NewHelper(log.With(logger, "module", "service.prom.chatgroup")),
		chatGroupBiz: chatGroupBiz,
		notifyBiz:    notifyBiz,
	}
}

func (s *ChatGroupService) CreateChatGroup(ctx context.Context, req *pb.CreateChatGroupRequest) (*pb.CreateChatGroupReply, error) {
	chatGroupBo := &bo.ChatGroupBO{
		Name:      req.GetName(),
		Remark:    req.GetRemark(),
		Hook:      req.GetHook(),
		NotifyApp: vobj.NotifyApp(req.GetApp()),
		HookName:  req.GetHookName(),
		Secret:    req.GetSecret(),
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
		Id:       req.GetId(),
		Name:     req.GetName(),
		Remark:   req.GetRemark(),
		HookName: req.GetHookName(),
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
	pgInfo := bo.NewPage(pgReq.GetCurr(), pgReq.GetSize())

	listReq := &bo.ListChatGroupReq{
		Page:    pgInfo,
		Keyword: req.GetKeyword(),
		Status:  vobj.Status(req.GetStatus()),
		App:     vobj.NotifyApp(req.GetApp()),
	}
	chatGroupBos, err := s.chatGroupBiz.ListChatGroup(ctx, listReq)
	if err != nil {
		s.log.Errorf("ListChatGroup err: %v", err)
		return nil, err
	}
	return &pb.ListChatGroupReply{
		Page: &api.PageReply{
			Curr:  pgInfo.GetRespCurr(),
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
	pgInfo := bo.NewPage(pgReq.GetCurr(), pgReq.GetSize())

	chatGroupBos, err := s.chatGroupBiz.ListChatGroup(ctx, &bo.ListChatGroupReq{
		Page:    pgInfo,
		Keyword: req.GetKeyword(),
		Status:  vobj.Status(req.GetStatus()),
	})
	if err != nil {
		s.log.Errorf("ListChatGroup err: %v", err)
		return nil, err
	}
	return &pb.SelectChatGroupReply{
		Page: &api.PageReply{
			Curr:  pgInfo.GetRespCurr(),
			Size:  pgInfo.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: slices.To(chatGroupBos, func(i *bo.ChatGroupBO) *api.ChatGroupSelectV1 {
			return i.ToSelectApi()
		}),
	}, nil
}
