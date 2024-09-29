package user

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/user"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/util/types"
)

type MessageService struct {
	pb.UnimplementedMessageServer

	userMessageBiz *biz.UserMessageBiz
}

func NewMessageService(userMessageBiz *biz.UserMessageBiz) *MessageService {
	return &MessageService{
		userMessageBiz: userMessageBiz,
	}
}

func (s *MessageService) DeleteMessages(ctx context.Context, req *pb.DeleteMessagesRequest) (*pb.DeleteMessagesReply, error) {
	if req.GetAll() {
		return &pb.DeleteMessagesReply{}, s.userMessageBiz.DeleteAllUserMessage(ctx)
	}
	if err := s.userMessageBiz.DeleteUserMessage(ctx, req.GetIds()); err != nil {
		return nil, err
	}
	return &pb.DeleteMessagesReply{}, nil
}

func (s *MessageService) ListMessage(ctx context.Context, req *pb.ListMessageRequest) (*pb.ListMessageReply, error) {
	params := &bo.QueryUserMessageListParams{
		Keyword: req.GetKeyword(),
		Page:    types.NewPagination(req.GetPagination()),
	}
	list, err := s.userMessageBiz.ListUserMessage(ctx, params)
	if err != nil {
		return nil, err
	}
	build := builder.NewParamsBuild().WithContext(ctx)
	return &pb.ListMessageReply{
		List:       build.UserModuleBuilder().NoticeUserMessageBuilder().ToAPIs(list),
		Pagination: build.PaginationModuleBuilder().ToAPI(params.Page),
	}, nil
}
