package team

import (
	"context"

	"github.com/aide-family/moon/api/admin"
	pb "github.com/aide-family/moon/api/admin/team"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/helper/model"
	"github.com/aide-family/moon/pkg/helper/model/bizmodel"
	"github.com/aide-family/moon/pkg/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type Service struct {
	pb.UnimplementedTeamServer

	teamBiz *biz.TeamBiz
}

func NewTeamService(teamBiz *biz.TeamBiz) *Service {
	return &Service{
		teamBiz: teamBiz,
	}
}

func (s *Service) CreateTeam(ctx context.Context, req *pb.CreateTeamRequest) (*pb.CreateTeamReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	leaderId := req.GetLeaderId()
	if leaderId <= 0 {
		leaderId = claims.GetUser()
	}
	params := &bo.CreateTeamParams{
		Name:      req.GetName(),
		Remark:    req.GetRemark(),
		CreatorID: claims.GetUser(),
		Logo:      req.GetLogo(),
		Status:    vobj.Status(req.GetStatus()),
		LeaderID:  leaderId,
		Admins:    req.GetAdminIds(),
	}
	_, err := s.teamBiz.CreateTeam(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.CreateTeamReply{}, nil
}

func (s *Service) UpdateTeam(ctx context.Context, req *pb.UpdateTeamRequest) (*pb.UpdateTeamReply, error) {
	params := &bo.UpdateTeamParams{
		ID:     req.GetId(),
		Name:   req.GetName(),
		Remark: req.GetRemark(),
		Logo:   req.GetLogo(),
		Status: vobj.Status(req.GetStatus()),
	}
	if err := s.teamBiz.UpdateTeam(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &pb.UpdateTeamReply{}, nil
}

func (s *Service) GetTeam(ctx context.Context, req *pb.GetTeamRequest) (*pb.GetTeamReply, error) {
	teamInfo, err := s.teamBiz.GetTeam(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.GetTeamReply{
		Team: build.NewTeamBuild(teamInfo).ToApi(ctx),
	}, nil
}

func (s *Service) ListTeam(ctx context.Context, req *pb.ListTeamRequest) (*pb.ListTeamReply, error) {
	params := &bo.QueryTeamListParams{
		Page:      types.NewPagination(req.GetPagination()),
		Keyword:   req.GetKeyword(),
		Status:    vobj.Status(req.GetStatus()),
		CreatorID: req.GetCreatorId(),
		LeaderID:  req.GetLeaderId(),
	}
	teamList, err := s.teamBiz.ListTeam(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.ListTeamReply{
		Pagination: build.NewPageBuild(params.Page).ToApi(),
		List: types.SliceTo(teamList, func(team *model.SysTeam) *admin.Team {
			return build.NewTeamBuild(team).ToApi(ctx)
		}),
	}, nil
}

func (s *Service) UpdateTeamStatus(ctx context.Context, req *pb.UpdateTeamStatusRequest) (*pb.UpdateTeamStatusReply, error) {
	if err := s.teamBiz.UpdateTeamStatus(ctx, vobj.Status(req.GetStatus()), req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &pb.UpdateTeamStatusReply{}, nil
}

func (s *Service) MyTeam(ctx context.Context, _ *pb.MyTeamRequest) (*pb.MyTeamReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	teamList, err := s.teamBiz.GetUserTeamList(ctx, claims.GetUser())
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.MyTeamReply{
		List: types.SliceTo(teamList, func(team *model.SysTeam) *admin.Team {
			return build.NewTeamBuild(team).ToApi(ctx)
		}),
	}, nil
}

func (s *Service) AddTeamMember(ctx context.Context, req *pb.AddTeamMemberRequest) (*pb.AddTeamMemberReply, error) {
	params := &bo.AddTeamMemberParams{
		ID: req.GetId(),
		Members: types.SliceTo(req.GetMembers(), func(member *pb.AddTeamMemberRequest_MemberItem) *bo.AddTeamMemberItem {
			return &bo.AddTeamMemberItem{
				UserID:  member.GetUserId(),
				Role:    vobj.Role(member.GetRole()),
				RoleIds: member.GetRoles(),
			}
		}),
	}
	if err := s.teamBiz.AddTeamMember(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &pb.AddTeamMemberReply{}, nil
}

func (s *Service) RemoveTeamMember(ctx context.Context, req *pb.RemoveTeamMemberRequest) (*pb.RemoveTeamMemberReply, error) {
	params := &bo.RemoveTeamMemberParams{
		ID:        req.GetId(),
		MemberIds: []uint32{req.GetUserId()},
	}
	if err := s.teamBiz.RemoveTeamMember(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &pb.RemoveTeamMemberReply{}, nil
}

func (s *Service) SetTeamAdmin(ctx context.Context, req *pb.SetTeamAdminRequest) (*pb.SetTeamAdminReply, error) {
	params := &bo.SetMemberAdminParams{
		ID:        req.GetId(),
		MemberIds: []uint32{req.GetUserId()},
		Role:      vobj.RoleAdmin,
	}
	if err := s.teamBiz.SetTeamAdmin(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &pb.SetTeamAdminReply{}, nil
}

func (s *Service) RemoveTeamAdmin(ctx context.Context, req *pb.RemoveTeamAdminRequest) (*pb.RemoveTeamAdminReply, error) {
	params := &bo.SetMemberAdminParams{
		ID:        req.GetId(),
		MemberIds: []uint32{req.GetUserId()},
		Role:      vobj.RoleUser,
	}
	if err := s.teamBiz.SetTeamAdmin(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &pb.RemoveTeamAdminReply{}, nil
}

func (s *Service) SetMemberRole(ctx context.Context, req *pb.SetMemberRoleRequest) (*pb.SetMemberRoleReply, error) {
	params := &bo.SetMemberRoleParams{
		ID:       req.GetId(),
		MemberID: req.GetUserId(),
		RoleIds:  req.GetRoles(),
	}
	if err := s.teamBiz.SetMemberRole(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &pb.SetMemberRoleReply{}, nil
}

func (s *Service) ListTeamMember(ctx context.Context, req *pb.ListTeamMemberRequest) (*pb.ListTeamMemberReply, error) {
	params := &bo.ListTeamMemberParams{
		Page:    types.NewPagination(req.GetPagination()),
		ID:      req.GetId(),
		Keyword: req.GetKeyword(),
		Role:    vobj.Role(req.GetRole()),
		Gender:  vobj.Gender(req.GetGender()),
		Status:  vobj.Status(req.GetStatus()),
	}
	memberList, err := s.teamBiz.ListTeamMember(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.ListTeamMemberReply{
		Pagination: build.NewPageBuild(params.Page).ToApi(),
		List: types.SliceTo(memberList, func(member *bizmodel.SysTeamMember) *admin.TeamMember {
			return build.NewTeamMemberBuild(member).ToApi(ctx)
		}),
	}, nil
}

func (s *Service) TransferTeamLeader(ctx context.Context, req *pb.TransferTeamLeaderRequest) (*pb.TransferTeamLeaderReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	params := &bo.TransferTeamLeaderParams{
		ID:          req.GetId(),
		LeaderID:    req.GetUserId(),
		OldLeaderID: claims.GetUser(),
	}
	if err := s.teamBiz.TransferTeamLeader(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &pb.TransferTeamLeaderReply{}, nil
}
