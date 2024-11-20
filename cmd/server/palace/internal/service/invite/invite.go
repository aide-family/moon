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

// NewInviteService 创建邀请操作服务
func NewInviteService(inviteBiz *biz.InviteBiz) *Service {
	return &Service{
		inviteBiz: inviteBiz,
	}
}

// InviteUser 邀请用户
func (s *Service) InviteUser(ctx context.Context, req *pb.InviteUserRequest) (*pb.InviteUserReply, error) {
	param := builder.NewParamsBuild(ctx).InviteModuleBuilder().WithInviteUserRequest(req).ToBo()
	err := s.inviteBiz.InviteUser(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.InviteUserReply{}, nil
}

// UpdateInviteStatus 更新邀请状态
func (s *Service) UpdateInviteStatus(ctx context.Context, req *pb.UpdateInviteStatusRequest) (*pb.UpdateInviteStatusReply, error) {
	param := builder.NewParamsBuild(ctx).InviteModuleBuilder().WithUpdateInviteStatusRequest(req).ToBo()
	err := s.inviteBiz.UpdateInviteStatus(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.UpdateInviteStatusReply{}, nil
}

// DeleteInvite 删除邀请
func (s *Service) DeleteInvite(ctx context.Context, req *pb.DeleteInviteRequest) (*pb.DeleteInviteReply, error) {
	err := s.inviteBiz.DeleteInvite(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.DeleteInviteReply{}, nil
}

// GetInvite 获取邀请详情
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
		Detail: builder.NewParamsBuild(ctx).InviteModuleBuilder().DoInviteBuilder(teamInfo).ToAPI(detail),
	}, nil
}

// UserInviteList 获取用户邀请列表
func (s *Service) UserInviteList(ctx context.Context, req *pb.ListUserInviteRequest) (*pb.ListUserInviteReply, error) {
	param := builder.NewParamsBuild(ctx).InviteModuleBuilder().WithListInviteUserRequest(req).ToBo()
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
		List:       builder.NewParamsBuild(ctx).InviteModuleBuilder().DoInviteBuilder(teamInfo).ToAPIs(list),
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(param.Page),
	}, nil
}
