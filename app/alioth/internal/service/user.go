package service

import (
	"context"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/user"
)

type UserService struct {
	v1.UnimplementedUserServiceServer
	directory *user.Directory
}

func NewUserService(directory *user.Directory) *UserService {
	return &UserService{directory: directory}
}

func (s *UserService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	account, err := s.directory.GetUser(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetUserResponse{User: userItem(*account)}, nil
}

func (s *UserService) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	items, total, page, pageSize, err := s.directory.ListUsers(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListUsersResponse{Items: userItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *UserService) SelectUsers(ctx context.Context, req *v1.SelectUsersRequest) (*v1.SelectUsersResponse, error) {
	items, err := s.directory.SelectUsers(ctx, req.GetLimit(), req.GetLastUid())
	if err != nil {
		return nil, err
	}
	return &v1.SelectUsersResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *UserService) BanUser(ctx context.Context, req *v1.BanUserRequest) (*v1.BanUserResponse, error) {
	account, err := s.directory.BanUser(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.BanUserResponse{User: userItem(*account)}, nil
}

func (s *UserService) PermitUser(ctx context.Context, req *v1.PermitUserRequest) (*v1.PermitUserResponse, error) {
	account, err := s.directory.PermitUser(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.PermitUserResponse{User: userItem(*account)}, nil
}

func (s *UserService) UpdateUserSystemRole(ctx context.Context, req *v1.UpdateUserSystemRoleRequest) (*v1.UpdateUserSystemRoleResponse, error) {
	account, err := s.directory.UpdateSystemRole(ctx, req.GetUid(), req.GetRole())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateUserSystemRoleResponse{User: userItem(*account)}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error) {
	if err := s.directory.DeleteUser(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteUserResponse{}, nil
}
