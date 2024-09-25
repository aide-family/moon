package service

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/invite"
)

// InviteService invite service
type InviteService struct {
	pb.UnimplementedInviteServer
}

// NewInviteService new a InviteService
func NewInviteService() *InviteService {
	return &InviteService{}
}

func (s *InviteService) InviteUser(ctx context.Context, req *pb.InviteUserRequest) (*pb.InviteUserReply, error) {
	return &pb.InviteUserReply{}, nil
}
func (s *InviteService) UpdateInviteStatus(ctx context.Context, req *pb.UpdateInviteStatusRequest) (*pb.UpdateInviteStatusReply, error) {
	return &pb.UpdateInviteStatusReply{}, nil
}
func (s *InviteService) DeleteInvite(ctx context.Context, req *pb.DeleteInviteRequest) (*pb.DeleteInviteReply, error) {
	return &pb.DeleteInviteReply{}, nil
}
func (s *InviteService) GetInvite(ctx context.Context, req *pb.GetInviteRequest) (*pb.GetInviteReply, error) {
	return &pb.GetInviteReply{}, nil
}
func (s *InviteService) ListInvite(ctx context.Context, req *pb.ListInviteRequest) (*pb.ListInviteReply, error) {
	return &pb.ListInviteReply{}, nil
}
