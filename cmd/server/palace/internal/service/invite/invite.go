package invite

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/invite"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
)

// Service invite service
type Service struct {
	pb.UnimplementedInviteServer

	inviteBiz *biz.InviteBiz
}

// NewInviteService new a InviteService
func NewInviteService(inviteBiz *biz.InviteBiz) *Service {
	return &Service{
		inviteBiz: inviteBiz,
	}
}

func (s *Service) InviteUser(ctx context.Context, req *pb.InviteUserRequest) (*pb.InviteUserReply, error) {
	param := builder.NewParamsBuild().WithContext(ctx).InviteModuleBuilder().WithInviteUserRequest(req).ToBo()
	err := s.inviteBiz.InviteUser(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.InviteUserReply{}, nil
}
func (s *Service) UpdateInviteStatus(ctx context.Context, req *pb.UpdateInviteStatusRequest) (*pb.UpdateInviteStatusReply, error) {
	param := builder.NewParamsBuild().WithContext(ctx).InviteModuleBuilder().WithUpdateInviteStatusRequest(req).ToBo()
	err := s.inviteBiz.UpdateInviteStatus(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.UpdateInviteStatusReply{}, nil
}
func (s *Service) DeleteInvite(ctx context.Context, req *pb.DeleteInviteRequest) (*pb.DeleteInviteReply, error) {
	err := s.inviteBiz.DeleteInvite(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.DeleteInviteReply{}, nil
}
func (s *Service) GetInvite(ctx context.Context, req *pb.GetInviteRequest) (*pb.GetInviteReply, error) {
	detail, err := s.inviteBiz.TeamInviteDetail(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	teamMap := s.inviteBiz.GetTeamMapByIds(ctx, []uint32{detail.TeamID})
	roles := s.inviteBiz.GetTeamRoles(ctx, detail.TeamID, detail.GetRolesIds())

	teamInfo := &bo.InviteTeamInfoParams{
		TeamMap:   teamMap,
		TeamRoles: roles,
	}
	return &pb.GetInviteReply{
		Detail: builder.NewParamsBuild().InviteModuleBuilder().DoInviteBuilder(teamInfo).ToAPI(detail),
	}, nil
}
func (s *Service) UserInviteList(ctx context.Context, req *pb.ListUserInviteRequest) (*pb.ListUserInviteReply, error) {
	param := builder.NewParamsBuild().WithContext(ctx).InviteModuleBuilder().WithListInviteUserRequest(req).ToBo()
	list, err := s.inviteBiz.UserInviteList(ctx, param)

	if !types.IsNil(err) {
		return nil, err
	}

	teamIds := types.SliceTo(list, func(item *model.SysTeamInvite) uint32 {
		return item.TeamID
	})
	teamMap := s.inviteBiz.GetTeamMapByIds(ctx, teamIds)

	teamInfo := &bo.InviteTeamInfoParams{
		TeamMap: teamMap,
	}

	return &pb.ListUserInviteReply{
		List:       builder.NewParamsBuild().InviteModuleBuilder().DoInviteBuilder(teamInfo).ToAPIs(list),
		Pagination: builder.NewParamsBuild().PaginationModuleBuilder().ToAPI(param.Page),
	}, nil
}
