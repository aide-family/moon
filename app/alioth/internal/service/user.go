package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/user"
)

type UserService struct {
	v1.UnimplementedUserServiceServer
	users *user.Usecase
}

func NewUserService(uc *user.Usecase) *UserService {
	return &UserService{users: uc}
}

func (s *UserService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	account, err := s.users.GetUser(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetUserResponse{User: userItem(*account)}, nil
}

func (s *UserService) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	items, total, page, pageSize, err := s.users.ListUsers(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListUsersResponse{Items: userItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *UserService) SelectUsers(ctx context.Context, req *v1.SelectUsersRequest) (*v1.SelectUsersResponse, error) {
	items, err := s.users.SelectUsers(ctx, req.GetLimit(), req.GetLastUid())
	if err != nil {
		return nil, err
	}
	return &v1.SelectUsersResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *UserService) BanUser(ctx context.Context, req *v1.BanUserRequest) (*v1.BanUserResponse, error) {
	account, err := s.users.BanUser(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.BanUserResponse{User: userItem(*account)}, nil
}

func (s *UserService) PermitUser(ctx context.Context, req *v1.PermitUserRequest) (*v1.PermitUserResponse, error) {
	account, err := s.users.PermitUser(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.PermitUserResponse{User: userItem(*account)}, nil
}
