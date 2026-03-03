package service

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/jwt"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil/cnst"

	"github.com/aide-family/goddess/internal/biz"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
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
	goddessv1.UnimplementedSelfServer

	userBiz      *biz.User
	memberBiz    *biz.Member
	namespaceBiz *biz.Namespace
	loginBiz     *biz.LoginBiz
}

func (s *SelfService) Info(ctx context.Context, _ *goddessv1.InfoRequest) (*goddessv1.UserItem, error) {
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

func (s *SelfService) Namespaces(ctx context.Context, _ *goddessv1.InfoRequest) (*goddessv1.NamespacesReply, error) {
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
	namespaces := make([]*goddessv1.NamespaceItem, 0, len(nsList))
	for _, ns := range nsList {
		namespaces = append(namespaces, ns.ToAPIV1NamespaceItem())
	}
	return &goddessv1.NamespacesReply{Namespaces: namespaces}, nil
}

func (s *SelfService) ChangeEmail(ctx context.Context, req *goddessv1.ChangeEmailRequest) (*goddessv1.ChangeReply, error) {
	userUID := contextx.GetUserUID(ctx)
	if userUID <= 0 {
		return nil, merr.ErrorUnauthorized("%s is required", cnst.HTTPHeaderAuthorization)
	}
	if err := s.userBiz.ChangeEmail(ctx, userUID, req.Email); err != nil {
		return nil, err
	}
	return &goddessv1.ChangeReply{Message: "email changed successfully"}, nil
}

func (s *SelfService) ChangeAvatar(ctx context.Context, req *goddessv1.ChangeAvatarRequest) (*goddessv1.ChangeReply, error) {
	userUID := contextx.GetUserUID(ctx)
	if userUID <= 0 {
		return nil, merr.ErrorUnauthorized("%s is required", cnst.HTTPHeaderAuthorization)
	}
	if err := s.userBiz.ChangeAvatar(ctx, userUID, req.Avatar); err != nil {
		return nil, err
	}
	return &goddessv1.ChangeReply{Message: "avatar changed successfully"}, nil
}

func (s *SelfService) ChangeRemark(ctx context.Context, req *goddessv1.ChangeRemarkRequest) (*goddessv1.ChangeReply, error) {
	userUID := contextx.GetUserUID(ctx)
	if userUID <= 0 {
		return nil, merr.ErrorUnauthorized("%s is required", cnst.HTTPHeaderAuthorization)
	}
	if err := s.userBiz.ChangeRemark(ctx, userUID, req.Remark); err != nil {
		return nil, err
	}
	return &goddessv1.ChangeReply{Message: "remark changed successfully"}, nil
}

func (s *SelfService) RefreshToken(ctx context.Context, _ *goddessv1.RefreshTokenRequest) (*goddessv1.RefreshTokenReply, error) {
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
	return &goddessv1.RefreshTokenReply{Token: token}, nil
}
