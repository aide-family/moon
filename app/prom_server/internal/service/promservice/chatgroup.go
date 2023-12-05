package promservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	pb "prometheus-manager/api/prom/notify"

	"prometheus-manager/app/prom_server/internal/biz"
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
	return &pb.CreateChatGroupReply{}, nil
}

func (s *ChatGroupService) UpdateChatGroup(ctx context.Context, req *pb.UpdateChatGroupRequest) (*pb.UpdateChatGroupReply, error) {
	return &pb.UpdateChatGroupReply{}, nil
}

func (s *ChatGroupService) DeleteChatGroup(ctx context.Context, req *pb.DeleteChatGroupRequest) (*pb.DeleteChatGroupReply, error) {
	return &pb.DeleteChatGroupReply{}, nil
}

func (s *ChatGroupService) GetChatGroup(ctx context.Context, req *pb.GetChatGroupRequest) (*pb.GetChatGroupReply, error) {
	return &pb.GetChatGroupReply{}, nil
}

func (s *ChatGroupService) ListChatGroup(ctx context.Context, req *pb.ListChatGroupRequest) (*pb.ListChatGroupReply, error) {
	return &pb.ListChatGroupReply{}, nil
}

func (s *ChatGroupService) SelectChatGroup(ctx context.Context, req *pb.SelectChatGroupRequest) (*pb.SelectChatGroupReply, error) {
	return &pb.SelectChatGroupReply{}, nil
}
