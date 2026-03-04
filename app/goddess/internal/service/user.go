package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func NewUserService(userBiz *biz.User) *UserService {
	return &UserService{
		userBiz: userBiz,
	}
}

type UserService struct {
	apiv1.UnimplementedUserServer

	userBiz *biz.User
}

func (s *UserService) GetUser(ctx context.Context, req *apiv1.GetUserRequest) (*apiv1.UserItem, error) {
	user, err := s.userBiz.GetUser(ctx, snowflake.ParseInt64(req.Uid))
	if err != nil {
		return nil, err
	}
	return user.ToAPIV1UserItem(), nil
}

func (s *UserService) ListUser(ctx context.Context, req *apiv1.ListUserRequest) (*apiv1.ListUserReply, error) {
	listBo := bo.NewListUserBo(req)
	page, err := s.userBiz.ListUser(ctx, listBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListUserReply(page), nil
}

func (s *UserService) SelectUser(ctx context.Context, req *apiv1.SelectUserRequest) (*apiv1.SelectUserReply, error) {
	selectBo := bo.NewSelectUserBo(req)
	result, err := s.userBiz.SelectUser(ctx, selectBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1SelectUserReply(result), nil
}

func (s *UserService) BanUser(ctx context.Context, req *apiv1.BanUserRequest) (*apiv1.BanUserReply, error) {
	if err := s.userBiz.BanUser(ctx, snowflake.ParseInt64(req.Uid)); err != nil {
		return nil, err
	}
	return &apiv1.BanUserReply{Message: "user has been banned"}, nil
}

func (s *UserService) PermitUser(ctx context.Context, req *apiv1.PermitUserRequest) (*apiv1.PermitUserReply, error) {
	if err := s.userBiz.PermitUser(ctx, snowflake.ParseInt64(req.Uid)); err != nil {
		return nil, err
	}
	return &apiv1.PermitUserReply{Message: "user has been permitted"}, nil
}

func (s *UserService) ValidateUser(ctx context.Context, userUID snowflake.ID) error {
	return s.userBiz.ValidateUser(ctx, userUID)
}
