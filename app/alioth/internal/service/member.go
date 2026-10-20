package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/member"
)

type MemberService struct {
	v1.UnimplementedMemberServiceServer
	members *member.Usecase
}

func NewMemberService(uc *member.Usecase) *MemberService {
	return &MemberService{members: uc}
}

func (s *MemberService) ListMembers(ctx context.Context, req *v1.ListMembersRequest) (*v1.ListMembersResponse, error) {
	items, total, page, pageSize, err := s.members.ListMembers(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListMembersResponse{Items: memberItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *MemberService) GetMember(ctx context.Context, req *v1.GetMemberRequest) (*v1.GetMemberResponse, error) {
	item, err := s.members.GetMember(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetMemberResponse{Member: memberItem(*item)}, nil
}

func (s *MemberService) ListUserMembers(ctx context.Context, req *v1.ListUserMembersRequest) (*v1.ListUserMembersResponse, error) {
	items, err := s.members.ListUserMembers(ctx, req.GetUserUid())
	if err != nil {
		return nil, err
	}
	return &v1.ListUserMembersResponse{Items: memberItems(items)}, nil
}

func (s *MemberService) SelectMembers(ctx context.Context, req *v1.SelectMembersRequest) (*v1.SelectMembersResponse, error) {
	items, err := s.members.SelectMembers(ctx, req.GetLimit(), req.GetLastUid())
	if err != nil {
		return nil, err
	}
	return &v1.SelectMembersResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *MemberService) InviteMember(ctx context.Context, req *v1.InviteMemberRequest) (*v1.InviteMemberResponse, error) {
	item, sent, err := s.members.InviteMember(ctx, req.GetEmail(), req.GetRoleUid())
	if err != nil {
		return nil, err
	}
	return &v1.InviteMemberResponse{Invitation: inviteItem(*item), EmailSent: sent}, nil
}

func (s *MemberService) ListMemberInvites(ctx context.Context, _ *v1.ListMemberInvitesRequest) (*v1.ListMemberInvitesResponse, error) {
	items, err := s.members.ListInvites(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListMemberInvitesResponse{Items: inviteItems(items)}, nil
}

func (s *MemberService) RevokeMemberInvite(ctx context.Context, req *v1.RevokeMemberInviteRequest) (*v1.RevokeMemberInviteResponse, error) {
	if err := s.members.RevokeInvite(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.RevokeMemberInviteResponse{}, nil
}

func (s *MemberService) DeleteMember(ctx context.Context, req *v1.DeleteMemberRequest) (*v1.DeleteMemberResponse, error) {
	if err := s.members.DeleteMember(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteMemberResponse{}, nil
}

func (s *MemberService) UpdateMemberStatus(ctx context.Context, req *v1.UpdateMemberStatusRequest) (*v1.UpdateMemberStatusResponse, error) {
	item, err := s.members.UpdateMemberStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateMemberStatusResponse{Member: memberItem(*item)}, nil
}

func (s *MemberService) UpdateMemberRoles(ctx context.Context, req *v1.UpdateMemberRolesRequest) (*v1.UpdateMemberRolesResponse, error) {
	item, err := s.members.UpdateMemberRoles(ctx, req.GetUid(), req.GetRoleUids())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateMemberRolesResponse{Member: memberItem(*item)}, nil
}
