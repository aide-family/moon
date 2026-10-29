package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/auth"
	"github.com/aide-family/moon/app/alioth/internal/biz/namespace"
	"github.com/aide-family/moon/app/alioth/internal/biz/role"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
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

func (s *SelfService) SendSelfChangeEmailCode(ctx context.Context, req *v1.SendSelfChangeEmailCodeRequest) (*v1.SendSelfChangeEmailCodeResponse, error) {
	if err := s.authenticator.SendSelfChangeEmailCode(ctx, req.GetEmail(), req.GetCaptchaId(), req.GetCaptchaCode()); err != nil {
		return nil, err
	}
	return &v1.SendSelfChangeEmailCodeResponse{}, nil
}

func (s *SelfService) ChangeEmail(ctx context.Context, req *v1.ChangeEmailRequest) (*v1.ChangeEmailResponse, error) {
	account, err := s.authenticator.ChangeEmail(ctx, req.GetEmail(), req.GetCode(), req.GetCaptchaId(), req.GetCaptchaCode())
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
	names, readOnly, err := s.authority.Mine(ctx)
	if err != nil {
		return nil, err
	}
	response := &v1.GetSelfPermissionsResponse{Permissions: names, ReadOnly: readOnly}
	// 只读切入时前端用 namespace 展示横幅；成员视角也一并返回，便于切换后刷新租户信息。
	current, err := s.tenancy.Current(ctx)
	if err != nil {
		return nil, err
	}
	if current != nil {
		response.Namespace = namespaceItem(*current)
	}
	return response, nil
}

func (s *SelfService) ListSelfOAuthLinks(ctx context.Context, _ *v1.ListSelfOAuthLinksRequest) (*v1.ListSelfOAuthLinksResponse, error) {
	links, err := s.authenticator.SelfOAuthLinks(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListSelfOAuthLinksResponse{Items: oauthLinkItems(links)}, nil
}

func (s *SelfService) BindSelfOAuth(ctx context.Context, req *v1.BindSelfOAuthRequest) (*v1.BindSelfOAuthResponse, error) {
	link, err := s.authenticator.BindSelfOAuth(ctx, req.GetProvider(), req.GetCode(), req.GetRedirectUri())
	if err != nil {
		return nil, err
	}
	return &v1.BindSelfOAuthResponse{Item: oauthLinkItem(*link)}, nil
}

func (s *SelfService) UnbindSelfOAuth(ctx context.Context, req *v1.UnbindSelfOAuthRequest) (*v1.UnbindSelfOAuthResponse, error) {
	if err := s.authenticator.UnbindSelfOAuth(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.UnbindSelfOAuthResponse{}, nil
}

// oauthLinkItem 把第三方绑定映射成接口返回。BoundAt 是绑定时间，按 RFC3339 输出；
// 时间为零时留空，避免前端渲染成 0001-01-01。
func oauthLinkItem(link bo.OAuthLink) *v1.OAuthLink {
	boundAt := ""
	if !link.BoundAt.IsZero() {
		boundAt = link.BoundAt.UTC().Format(time.RFC3339)
	}
	return &v1.OAuthLink{
		Uid: link.UID, Provider: link.Provider, OpenId: link.OpenID, Email: link.Email, BoundAt: boundAt,
	}
}

func oauthLinkItems(items []bo.OAuthLink) []*v1.OAuthLink {
	out := make([]*v1.OAuthLink, 0, len(items))
	for _, item := range items {
		out = append(out, oauthLinkItem(item))
	}
	return out
}
