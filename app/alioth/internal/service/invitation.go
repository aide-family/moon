package service

import (
	"context"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/member"
)

// InvitationService 是平台邀请消息入口。查看邀请函不需要登录，确认需要登录。
type InvitationService struct {
	v1.UnimplementedInvitationServiceServer
	membership *member.Membership
}

func NewInvitationService(membership *member.Membership) *InvitationService {
	return &InvitationService{membership: membership}
}

func (s *InvitationService) ListSelfInvitations(ctx context.Context, _ *v1.ListSelfInvitationsRequest) (*v1.ListSelfInvitationsResponse, error) {
	items, err := s.membership.ListSelfInvites(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListSelfInvitationsResponse{Items: inviteItems(items)}, nil
}

func (s *InvitationService) GetInvitation(ctx context.Context, req *v1.GetInvitationRequest) (*v1.GetInvitationResponse, error) {
	item, err := s.membership.GetInviteByToken(ctx, req.GetToken())
	if err != nil {
		return nil, err
	}
	return &v1.GetInvitationResponse{Invitation: inviteItem(*item)}, nil
}

func (s *InvitationService) AcceptInvitation(ctx context.Context, req *v1.AcceptInvitationRequest) (*v1.AcceptInvitationResponse, error) {
	item, err := s.membership.AcceptInvite(ctx, req.GetToken(), req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.AcceptInvitationResponse{Member: memberItem(*item)}, nil
}

func (s *InvitationService) DeclineInvitation(ctx context.Context, req *v1.DeclineInvitationRequest) (*v1.DeclineInvitationResponse, error) {
	if err := s.membership.DeclineInvite(ctx, req.GetToken(), req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeclineInvitationResponse{}, nil
}
