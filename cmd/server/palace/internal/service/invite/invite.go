package invite

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/invite"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/util/types"
)

// Service invite service
type Service struct {
	pb.UnimplementedInviteServer

	inviteBiz *biz.InviteBiz
}

// NewInviteService new a InviteService
func NewInviteService() *Service {
	return &Service{}
}

func (s *Service) InviteUser(ctx context.Context, req *pb.InviteUserRequest) (*pb.InviteUserReply, error) {
	param := builder.NewParamsBuild().InviteModuleBuilder().WithInviteUserRequest(req).ToBo()
	err := s.inviteBiz.InviteUser(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.InviteUserReply{}, nil
}
func (s *Service) UpdateInviteStatus(ctx context.Context, req *pb.UpdateInviteStatusRequest) (*pb.UpdateInviteStatusReply, error) {
	param := builder.NewParamsBuild().InviteModuleBuilder().WithUpdateInviteStatusRequest(req).ToBo()
	err := s.inviteBiz.UpdateInviteStatus(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.UpdateInviteStatusReply{}, nil
}
func (s *Service) DeleteInvite(ctx context.Context, req *pb.DeleteInviteRequest) (*pb.DeleteInviteReply, error) {
	return &pb.DeleteInviteReply{}, nil
}
func (s *Service) GetInvite(ctx context.Context, req *pb.GetInviteRequest) (*pb.GetInviteReply, error) {
	return &pb.GetInviteReply{}, nil
}
func (s *Service) ListInvite(ctx context.Context, req *pb.ListInviteRequest) (*pb.ListInviteReply, error) {
	param := builder.NewParamsBuild().InviteModuleBuilder().WithListInviteUserRequest(req).ToBo()
	_, err := s.inviteBiz.InviteList(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.ListInviteReply{}, nil
}
