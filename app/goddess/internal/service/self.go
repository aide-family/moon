package service

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/jwt"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil/cnst"

	"github.com/aide-family/goddess/internal/biz"
	magicboxv1 "github.com/aide-family/magicbox/api/v1"
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
	magicboxv1.UnimplementedSelfServer

	userBiz      *biz.User
	memberBiz    *biz.Member
	namespaceBiz *biz.Namespace
	loginBiz     *biz.LoginBiz
}

func (s *SelfService) Info(ctx context.Context, _ *magicboxv1.InfoRequest) (*magicboxv1.UserItem, error) {
	userUID := contextx.GetUserUID(ctx)
	if userUID <= 0 {
		return nil, merr.ErrorUnauthorized("%s is required", cnst.HTTPHeaderAuthorization)
	}
	user, err := s.userBiz.GetUser(ctx, userUID)
	if err != nil {
		return nil, err
	}
	return user.ToAPIV1UserItem(), nil
}

func (s *SelfService) Namespaces(ctx context.Context, _ *magicboxv1.InfoRequest) (*magicboxv1.NamespacesReply, error) {
	userUID := contextx.GetUserUID(ctx)
	if userUID <= 0 {
		return nil, merr.ErrorUnauthorized("%s is required", cnst.HTTPHeaderAuthorization)
	}
	namespaceUIDs, err := s.memberBiz.GetNamespaceUIDsByUserUID(ctx, userUID)
	if err != nil {
		return nil, err
	}
	nsList, err := s.namespaceBiz.ListNamespacesByUIDs(ctx, namespaceUIDs)
	if err != nil {
		return nil, err
	}
	namespaces := make([]*magicboxv1.NamespaceItem, 0, len(nsList))
	for _, ns := range nsList {
		namespaces = append(namespaces, ns.ToAPIV1NamespaceItem())
	}
	return &magicboxv1.NamespacesReply{Namespaces: namespaces}, nil
}

func (s *SelfService) ChangeEmail(ctx context.Context, req *magicboxv1.ChangeEmailRequest) (*magicboxv1.ChangeEmailReply, error) {
	userUID := contextx.GetUserUID(ctx)
	if userUID <= 0 {
		return nil, merr.ErrorUnauthorized("%s is required", cnst.HTTPHeaderAuthorization)
	}
	if err := s.userBiz.ChangeEmail(ctx, userUID, req.Email); err != nil {
		return nil, err
	}
	return &magicboxv1.ChangeEmailReply{Message: "email changed successfully"}, nil
}

func (s *SelfService) ChangeAvatar(ctx context.Context, req *magicboxv1.ChangeAvatarRequest) (*magicboxv1.ChangeAvatarReply, error) {
	userUID := contextx.GetUserUID(ctx)
	if userUID <= 0 {
		return nil, merr.ErrorUnauthorized("%s is required", cnst.HTTPHeaderAuthorization)
	}
	if err := s.userBiz.ChangeAvatar(ctx, userUID, req.Avatar); err != nil {
		return nil, err
	}
	return &magicboxv1.ChangeAvatarReply{Message: "avatar changed successfully"}, nil
}

func (s *SelfService) RefreshToken(ctx context.Context, _ *magicboxv1.RefreshTokenRequest) (*magicboxv1.RefreshTokenReply, error) {
	userUID := contextx.GetUserUID(ctx)
	if userUID <= 0 {
		return nil, merr.ErrorUnauthorized("%s is required", cnst.HTTPHeaderAuthorization)
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
	return &magicboxv1.RefreshTokenReply{Token: token}, nil
}
