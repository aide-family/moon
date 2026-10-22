package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/auth"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/namespace"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/role"
)

type SelfService struct {
	v1.UnimplementedSelfServiceServer
	authenticator *auth.Authenticator
	tenancy       *namespace.Tenancy
	authority     *role.Authority
}

func NewSelfService(authenticator *auth.Authenticator, tenancy *namespace.Tenancy, authority *role.Authority) *SelfService {
	return &SelfService{authenticator: authenticator, tenancy: tenancy, authority: authority}
}

func (s *SelfService) GetSelfInfo(ctx context.Context, _ *v1.GetSelfInfoRequest) (*v1.GetSelfInfoResponse, error) {
	account, err := s.authenticator.Self(ctx)
	if err != nil {
		return nil, err
	}
	role, permissions, err := s.authenticator.SelfSystem(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSelfInfoResponse{
		User:              userItem(*account),
		SystemRole:        role,
		SystemPermissions: permissions,
	}, nil
}

func (s *SelfService) ListSelfNamespaces(ctx context.Context, _ *v1.ListSelfNamespacesRequest) (*v1.ListSelfNamespacesResponse, error) {
	items, err := s.tenancy.ListMine(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListSelfNamespacesResponse{Items: namespaceItems(items)}, nil
}

func (s *SelfService) RefreshToken(ctx context.Context, _ *v1.RefreshTokenRequest) (*v1.RefreshTokenResponse, error) {
	token, err := s.authenticator.Refresh(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.RefreshTokenResponse{AccessToken: token}, nil
}

func (s *SelfService) ChangeEmail(ctx context.Context, req *v1.ChangeEmailRequest) (*v1.ChangeEmailResponse, error) {
	account, err := s.authenticator.ChangeEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, err
	}
	return &v1.ChangeEmailResponse{User: userItem(*account)}, nil
}

func (s *SelfService) ChangeAvatar(ctx context.Context, req *v1.ChangeAvatarRequest) (*v1.ChangeAvatarResponse, error) {
	account, err := s.authenticator.ChangeAvatar(ctx, req.GetAvatar())
	if err != nil {
		return nil, err
	}
	return &v1.ChangeAvatarResponse{User: userItem(*account)}, nil
}

func (s *SelfService) ChangeRemark(ctx context.Context, req *v1.ChangeRemarkRequest) (*v1.ChangeRemarkResponse, error) {
	account, err := s.authenticator.ChangeRemark(ctx, req.GetRemark())
	if err != nil {
		return nil, err
	}
	return &v1.ChangeRemarkResponse{User: userItem(*account)}, nil
}

func (s *SelfService) ChangePhone(ctx context.Context, req *v1.ChangePhoneRequest) (*v1.ChangePhoneResponse, error) {
	account, err := s.authenticator.ChangePhone(ctx, req.GetPhone())
	if err != nil {
		return nil, err
	}
	return &v1.ChangePhoneResponse{User: userItem(*account)}, nil
}

func (s *SelfService) GetSelfPermissions(ctx context.Context, _ *v1.GetSelfPermissionsRequest) (*v1.GetSelfPermissionsResponse, error) {
	names, err := s.authority.Mine(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSelfPermissionsResponse{Permissions: names}, nil
}
