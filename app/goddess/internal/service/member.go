package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func NewMemberService(memberBiz *biz.Member) *MemberService {
	return &MemberService{
		memberBiz: memberBiz,
	}
}

type MemberService struct {
	apiv1.UnimplementedMemberServer

	memberBiz *biz.Member
}

func (s *MemberService) InviteMember(ctx context.Context, req *apiv1.InviteMemberRequest) (*apiv1.InviteMemberReply, error) {
	inviteBo := &bo.InviteMemberBo{
		Email: req.Email,
		Role:  req.Role,
	}
	if err := s.memberBiz.InviteMember(ctx, inviteBo); err != nil {
		return nil, err
	}
	return &apiv1.InviteMemberReply{Message: "invitation sent successfully"}, nil
}

func (s *MemberService) DismissMember(ctx context.Context, req *apiv1.DismissMemberRequest) (*apiv1.DismissMemberReply, error) {
	if err := s.memberBiz.DismissMember(ctx, snowflake.ParseInt64(req.Uid)); err != nil {
		return nil, err
	}
	return &apiv1.DismissMemberReply{Message: "member dismissed successfully"}, nil
}

func (s *MemberService) GetMember(ctx context.Context, req *apiv1.GetMemberRequest) (*apiv1.MemberItem, error) {
	member, err := s.memberBiz.GetMember(ctx, snowflake.ParseInt64(req.Uid))
	if err != nil {
		return nil, err
	}
	return member.ToAPIV1MemberItem(), nil
}

func (s *MemberService) ListMember(ctx context.Context, req *apiv1.ListMemberRequest) (*apiv1.ListMemberReply, error) {
	listBo := bo.NewListMemberBo(req)
	page, err := s.memberBiz.ListMember(ctx, listBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListMemberReply(page), nil
}

func (s *MemberService) SelectMember(ctx context.Context, req *apiv1.SelectMemberRequest) (*apiv1.SelectMemberReply, error) {
	selectBo := bo.NewSelectMemberBo(req)
	result, err := s.memberBiz.SelectMember(ctx, selectBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1SelectMemberReply(result), nil
}

func (s *MemberService) UpdateMemberStatus(ctx context.Context, req *apiv1.UpdateMemberStatusRequest) (*apiv1.UpdateMemberStatusReply, error) {
	if err := s.memberBiz.UpdateMemberStatus(ctx, snowflake.ParseInt64(req.Uid), int32(req.Status)); err != nil {
		return nil, err
	}
	return &apiv1.UpdateMemberStatusReply{Message: "status updated successfully"}, nil
}
