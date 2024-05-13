package user

import (
	"context"

	pb "github.com/aide-cloud/moon/api/admin/user"
	"github.com/aide-cloud/moon/cmd/moon/internal/biz"
)

type Service struct {
	pb.UnimplementedUserServer

	userBiz *biz.UserBiz
}

func NewUserService(userBiz *biz.UserBiz) *Service {
	return &Service{
		userBiz: userBiz,
	}
}

func (s *Service) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserReply, error) {
	return &pb.CreateUserReply{}, nil
}
func (s *Service) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserReply, error) {
	return &pb.UpdateUserReply{}, nil
}
func (s *Service) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserReply, error) {
	return &pb.DeleteUserReply{}, nil
}
func (s *Service) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserReply, error) {
	return &pb.GetUserReply{}, nil
}
func (s *Service) ListUser(ctx context.Context, req *pb.ListUserRequest) (*pb.ListUserReply, error) {
	return &pb.ListUserReply{}, nil
}
func (s *Service) BatchUpdateUserStatus(ctx context.Context, req *pb.BatchUpdateUserStatusRequest) (*pb.BatchUpdateUserStatusReply, error) {
	return &pb.BatchUpdateUserStatusReply{}, nil
}
func (s *Service) ResetUserPassword(ctx context.Context, req *pb.ResetUserPasswordRequest) (*pb.ResetUserPasswordReply, error) {
	return &pb.ResetUserPasswordReply{}, nil
}
func (s *Service) ResetUserPasswordBySelf(ctx context.Context, req *pb.ResetUserPasswordBySelfRequest) (*pb.ResetUserPasswordBySelfReply, error) {
	return &pb.ResetUserPasswordBySelfReply{}, nil
}
func (s *Service) GetUserSelectList(ctx context.Context, req *pb.GetUserSelectListRequest) (*pb.GetUserSelectListReply, error) {
	return &pb.GetUserSelectListReply{}, nil
}
