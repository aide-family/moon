package service

import (
	"context"

	"github.com/aide-family/magicbox/jwt"

	"github.com/aide-family/goddess/internal/biz"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func NewSelfService(userBiz *biz.User, memberBiz *biz.Member, namespaceBiz *biz.Namespace, loginBiz *biz.LoginBiz) *SelfService {
	return &SelfService{
		userBiz:      userBiz,
		memberBiz:    memberBiz,
		namespaceBiz: namespaceBiz,
		loginBiz:     loginBiz,
	}
}

type SelfService struct {
	apiv1.UnimplementedSelfServer

	userBiz      *biz.User
	memberBiz    *biz.Member
	namespaceBiz *biz.Namespace
	loginBiz     *biz.LoginBiz
}

func (s *SelfService) Info(ctx context.Context, _ *apiv1.InfoRequest) (*apiv1.UserItem, error) {
	userUID, err := s.userBiz.GetUserUID(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.userBiz.GetUser(ctx, userUID)
	if err != nil {
		return nil, err
	}
	return user.ToAPIV1UserItem(), nil
}

func (s *SelfService) Namespaces(ctx context.Context, _ *apiv1.InfoRequest) (*apiv1.NamespacesReply, error) {
	userUID, err := s.userBiz.GetUserUID(ctx)
	if err != nil {
		return nil, err
	}
	namespaceUIDs, err := s.memberBiz.GetNamespaceUIDsByUserUID(ctx, userUID)
	if err != nil {
		return nil, err
	}
	nsList, err := s.namespaceBiz.ListNamespacesByUIDs(ctx, namespaceUIDs)
	if err != nil {
		return nil, err
	}
	namespaces := make([]*apiv1.NamespaceItem, 0, len(nsList))
	for _, ns := range nsList {
		namespaces = append(namespaces, ns.ToAPIV1NamespaceItem())
	}
	return &apiv1.NamespacesReply{Namespaces: namespaces}, nil
}

func (s *SelfService) ChangeEmail(ctx context.Context, req *apiv1.ChangeEmailRequest) (*apiv1.ChangeReply, error) {
	userUID, err := s.userBiz.GetUserUID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.userBiz.ChangeEmail(ctx, userUID, req.Email); err != nil {
		return nil, err
	}
	return &apiv1.ChangeReply{Message: "email changed successfully"}, nil
}

func (s *SelfService) ChangeAvatar(ctx context.Context, req *apiv1.ChangeAvatarRequest) (*apiv1.ChangeReply, error) {
	userUID, err := s.userBiz.GetUserUID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.userBiz.ChangeAvatar(ctx, userUID, req.Avatar); err != nil {
		return nil, err
	}
	return &apiv1.ChangeReply{Message: "avatar changed successfully"}, nil
}

func (s *SelfService) ChangeRemark(ctx context.Context, req *apiv1.ChangeRemarkRequest) (*apiv1.ChangeReply, error) {
	userUID, err := s.userBiz.GetUserUID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.userBiz.ChangeRemark(ctx, userUID, req.Remark); err != nil {
		return nil, err
	}
	return &apiv1.ChangeReply{Message: "remark changed successfully"}, nil
}

func (s *SelfService) RefreshToken(ctx context.Context, _ *apiv1.RefreshTokenRequest) (*apiv1.RefreshTokenReply, error) {
	userUID, err := s.userBiz.GetUserUID(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.userBiz.GetUser(ctx, userUID)
	if err != nil {
		return nil, err
	}
	token, err := s.loginBiz.RefreshToken(ctx, jwt.BaseInfo{
		UID:      user.UID,
		Username: user.Name,
	})
	if err != nil {
		return nil, err
	}
	return &apiv1.RefreshTokenReply{Token: token}, nil
}
