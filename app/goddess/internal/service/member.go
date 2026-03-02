package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func NewMemberService(memberBiz *biz.Member) *MemberService {
	return &MemberService{
		memberBiz: memberBiz,
	}
}

type MemberService struct {
	goddessv1.UnimplementedMemberServer

	memberBiz *biz.Member
}

func (s *MemberService) InviteMember(ctx context.Context, req *goddessv1.InviteMemberRequest) (*goddessv1.InviteMemberReply, error) {
	inviteBo := &bo.InviteMemberBo{
		Email:   req.Email,
		RoleUID: req.RoleUID,
	}
	if err := s.memberBiz.InviteMember(ctx, inviteBo); err != nil {
		return nil, err
	}
	return &goddessv1.InviteMemberReply{Message: "invitation sent successfully"}, nil
}

func (s *MemberService) DismissMember(ctx context.Context, req *goddessv1.DismissMemberRequest) (*goddessv1.DismissMemberReply, error) {
	if err := s.memberBiz.DismissMember(ctx, snowflake.ParseInt64(req.Uid)); err != nil {
		return nil, err
	}
	return &goddessv1.DismissMemberReply{Message: "member dismissed successfully"}, nil
}

func (s *MemberService) GetMember(ctx context.Context, req *goddessv1.GetMemberRequest) (*goddessv1.MemberItem, error) {
	member, err := s.memberBiz.GetMember(ctx, snowflake.ParseInt64(req.Uid))
	if err != nil {
		return nil, err
	}
	return member.ToAPIV1MemberItem(), nil
}

func (s *MemberService) ListMember(ctx context.Context, req *goddessv1.ListMemberRequest) (*goddessv1.ListMemberReply, error) {
	listBo := bo.NewListMemberBo(req)
	page, err := s.memberBiz.ListMember(ctx, listBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListMemberReply(page), nil
}

func (s *MemberService) SelectMember(ctx context.Context, req *goddessv1.SelectMemberRequest) (*goddessv1.SelectMemberReply, error) {
	selectBo := bo.NewSelectMemberBo(req)
	result, err := s.memberBiz.SelectMember(ctx, selectBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1SelectMemberReply(result), nil
}

func (s *MemberService) UpdateMemberStatus(ctx context.Context, req *goddessv1.UpdateMemberStatusRequest) (*goddessv1.UpdateMemberStatusReply, error) {
	if err := s.memberBiz.UpdateMemberStatus(ctx, snowflake.ParseInt64(req.Uid), int32(req.Status)); err != nil {
		return nil, err
	}
	return &goddessv1.UpdateMemberStatusReply{Message: "status updated successfully"}, nil
}
