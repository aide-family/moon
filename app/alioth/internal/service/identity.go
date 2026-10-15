package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz"
)

type IdentityService struct {
	v1.UnimplementedCaptchaServiceServer
	v1.UnimplementedAuthServiceServer
	v1.UnimplementedSelfServiceServer
	v1.UnimplementedNamespaceServiceServer
	v1.UnimplementedMemberServiceServer
	v1.UnimplementedUserServiceServer
	v1.UnimplementedRoleServiceServer
	uc *biz.IdentityUsecase
}

func NewIdentityService(uc *biz.IdentityUsecase) *IdentityService {
	return &IdentityService{uc: uc}
}

func (s *IdentityService) GetCaptcha(ctx context.Context, _ *v1.GetCaptchaRequest) (*v1.GetCaptchaResponse, error) {
	id, image, err := s.uc.GetCaptcha(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetCaptchaResponse{Id: id, Image: image}, nil
}

func (s *IdentityService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	session, err := s.uc.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetName(), req.GetCaptchaId(), req.GetCaptchaCode())
	if err != nil {
		return nil, err
	}
	return &v1.RegisterResponse{Session: tokenPair(session)}, nil
}

func (s *IdentityService) PasswordLogin(ctx context.Context, req *v1.PasswordLoginRequest) (*v1.PasswordLoginResponse, error) {
	session, err := s.uc.PasswordLogin(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}
	return &v1.PasswordLoginResponse{Session: tokenPair(session)}, nil
}

func (s *IdentityService) SendEmailLoginCode(ctx context.Context, req *v1.SendEmailLoginCodeRequest) (*v1.SendEmailLoginCodeResponse, error) {
	if err := s.uc.SendEmailLoginCode(ctx, req.GetEmail()); err != nil {
		return nil, err
	}
	return &v1.SendEmailLoginCodeResponse{}, nil
}

func (s *IdentityService) EmailLogin(ctx context.Context, req *v1.EmailLoginRequest) (*v1.EmailLoginResponse, error) {
	session, err := s.uc.EmailLogin(ctx, req.GetEmail(), req.GetCode())
	if err != nil {
		return nil, err
	}
	return &v1.EmailLoginResponse{Session: tokenPair(session)}, nil
}

func (s *IdentityService) OAuth2Login(ctx context.Context, req *v1.OAuth2LoginRequest) (*v1.OAuth2LoginResponse, error) {
	session, err := s.uc.OAuthLogin(ctx, req.GetProvider(), req.GetCode(), req.GetRedirectUri())
	if err != nil {
		return nil, err
	}
	return &v1.OAuth2LoginResponse{Session: tokenPair(session)}, nil
}

func (s *IdentityService) GetSelfInfo(ctx context.Context, _ *v1.GetSelfInfoRequest) (*v1.GetSelfInfoResponse, error) {
	account, err := s.uc.Self(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSelfInfoResponse{User: userItem(*account)}, nil
}

func (s *IdentityService) ListSelfNamespaces(ctx context.Context, _ *v1.ListSelfNamespacesRequest) (*v1.ListSelfNamespacesResponse, error) {
	items, err := s.uc.SelfNamespaces(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListSelfNamespacesResponse{Items: namespaceItems(items)}, nil
}

func (s *IdentityService) RefreshToken(ctx context.Context, _ *v1.RefreshTokenRequest) (*v1.RefreshTokenResponse, error) {
	token, err := s.uc.Refresh(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.RefreshTokenResponse{AccessToken: token}, nil
}

func (s *IdentityService) ChangeEmail(ctx context.Context, req *v1.ChangeEmailRequest) (*v1.ChangeEmailResponse, error) {
	account, err := s.uc.ChangeEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, err
	}
	return &v1.ChangeEmailResponse{User: userItem(*account)}, nil
}

func (s *IdentityService) ChangeAvatar(ctx context.Context, req *v1.ChangeAvatarRequest) (*v1.ChangeAvatarResponse, error) {
	account, err := s.uc.ChangeAvatar(ctx, req.GetAvatar())
	if err != nil {
		return nil, err
	}
	return &v1.ChangeAvatarResponse{User: userItem(*account)}, nil
}

func (s *IdentityService) ChangeRemark(ctx context.Context, req *v1.ChangeRemarkRequest) (*v1.ChangeRemarkResponse, error) {
	account, err := s.uc.ChangeRemark(ctx, req.GetRemark())
	if err != nil {
		return nil, err
	}
	return &v1.ChangeRemarkResponse{User: userItem(*account)}, nil
}

func (s *IdentityService) GetSelfPermissions(ctx context.Context, _ *v1.GetSelfPermissionsRequest) (*v1.GetSelfPermissionsResponse, error) {
	names, err := s.uc.SelfPermissions(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSelfPermissionsResponse{Permissions: names}, nil
}

func (s *IdentityService) CreateNamespace(ctx context.Context, req *v1.CreateNamespaceRequest) (*v1.CreateNamespaceResponse, error) {
	item, err := s.uc.CreateNamespace(ctx, req.GetName(), req.GetRemark())
	if err != nil {
		return nil, err
	}
	return &v1.CreateNamespaceResponse{Namespace: namespaceItem(*item)}, nil
}

func (s *IdentityService) UpdateNamespace(ctx context.Context, req *v1.UpdateNamespaceRequest) (*v1.UpdateNamespaceResponse, error) {
	item, err := s.uc.UpdateNamespace(ctx, req.GetUid(), req.GetName(), req.GetRemark(), req.GetLogo())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNamespaceResponse{Namespace: namespaceItem(*item)}, nil
}

func (s *IdentityService) UpdateNamespaceStatus(ctx context.Context, req *v1.UpdateNamespaceStatusRequest) (*v1.UpdateNamespaceStatusResponse, error) {
	item, err := s.uc.UpdateNamespaceStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNamespaceStatusResponse{Namespace: namespaceItem(*item)}, nil
}

func (s *IdentityService) DeleteNamespace(ctx context.Context, req *v1.DeleteNamespaceRequest) (*v1.DeleteNamespaceResponse, error) {
	if err := s.uc.DeleteNamespace(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteNamespaceResponse{}, nil
}

func (s *IdentityService) GetNamespace(ctx context.Context, req *v1.GetNamespaceRequest) (*v1.GetNamespaceResponse, error) {
	item, err := s.uc.GetNamespace(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetNamespaceResponse{Namespace: namespaceItem(*item)}, nil
}

func (s *IdentityService) ListNamespaces(ctx context.Context, req *v1.ListNamespacesRequest) (*v1.ListNamespacesResponse, error) {
	items, total, page, pageSize, err := s.uc.ListNamespaces(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListNamespacesResponse{Items: namespaceItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *IdentityService) SelectNamespaces(ctx context.Context, req *v1.SelectNamespacesRequest) (*v1.SelectNamespacesResponse, error) {
	items, err := s.uc.SelectNamespaces(ctx, req.GetLimit(), req.GetLastUid())
	if err != nil {
		return nil, err
	}
	return &v1.SelectNamespacesResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *IdentityService) ListSimpleNamespaces(ctx context.Context, req *v1.ListSimpleNamespacesRequest) (*v1.ListSimpleNamespacesResponse, error) {
	items, total, _, _, err := s.uc.ListNamespaces(ctx, req.GetPage(), req.GetPageSize(), "")
	if err != nil {
		return nil, err
	}
	return &v1.ListSimpleNamespacesResponse{Items: namespaceItems(items), Total: int32(total)}, nil
}

func (s *IdentityService) TransferNamespaceOwner(ctx context.Context, req *v1.TransferNamespaceOwnerRequest) (*v1.TransferNamespaceOwnerResponse, error) {
	item, err := s.uc.TransferOwner(ctx, req.GetUid(), req.GetUserUid())
	if err != nil {
		return nil, err
	}
	return &v1.TransferNamespaceOwnerResponse{Namespace: namespaceItem(*item)}, nil
}

func (s *IdentityService) ListMembers(ctx context.Context, req *v1.ListMembersRequest) (*v1.ListMembersResponse, error) {
	items, total, page, pageSize, err := s.uc.ListMembers(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListMembersResponse{Items: memberItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *IdentityService) GetMember(ctx context.Context, req *v1.GetMemberRequest) (*v1.GetMemberResponse, error) {
	item, err := s.uc.GetMember(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetMemberResponse{Member: memberItem(*item)}, nil
}

func (s *IdentityService) ListUserMembers(ctx context.Context, req *v1.ListUserMembersRequest) (*v1.ListUserMembersResponse, error) {
	items, err := s.uc.ListUserMembers(ctx, req.GetUserUid())
	if err != nil {
		return nil, err
	}
	return &v1.ListUserMembersResponse{Items: memberItems(items)}, nil
}

func (s *IdentityService) SelectMembers(ctx context.Context, req *v1.SelectMembersRequest) (*v1.SelectMembersResponse, error) {
	items, err := s.uc.SelectMembers(ctx, req.GetLimit(), req.GetLastUid())
	if err != nil {
		return nil, err
	}
	return &v1.SelectMembersResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *IdentityService) InviteMember(ctx context.Context, req *v1.InviteMemberRequest) (*v1.InviteMemberResponse, error) {
	item, err := s.uc.InviteMember(ctx, req.GetEmail(), req.GetRoleUid())
	if err != nil {
		return nil, err
	}
	return &v1.InviteMemberResponse{Member: memberItem(*item)}, nil
}

func (s *IdentityService) DeleteMember(ctx context.Context, req *v1.DeleteMemberRequest) (*v1.DeleteMemberResponse, error) {
	if err := s.uc.DeleteMember(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteMemberResponse{}, nil
}

func (s *IdentityService) UpdateMemberStatus(ctx context.Context, req *v1.UpdateMemberStatusRequest) (*v1.UpdateMemberStatusResponse, error) {
	item, err := s.uc.UpdateMemberStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateMemberStatusResponse{Member: memberItem(*item)}, nil
}

func (s *IdentityService) UpdateMemberRoles(ctx context.Context, req *v1.UpdateMemberRolesRequest) (*v1.UpdateMemberRolesResponse, error) {
	item, err := s.uc.UpdateMemberRoles(ctx, req.GetUid(), req.GetRoleUids())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateMemberRolesResponse{Member: memberItem(*item)}, nil
}

func (s *IdentityService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	account, err := s.uc.GetUser(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetUserResponse{User: userItem(*account)}, nil
}

func (s *IdentityService) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	items, total, page, pageSize, err := s.uc.ListUsers(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListUsersResponse{Items: userItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *IdentityService) SelectUsers(ctx context.Context, req *v1.SelectUsersRequest) (*v1.SelectUsersResponse, error) {
	items, err := s.uc.SelectUsers(ctx, req.GetLimit(), req.GetLastUid())
	if err != nil {
		return nil, err
	}
	return &v1.SelectUsersResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *IdentityService) BanUser(ctx context.Context, req *v1.BanUserRequest) (*v1.BanUserResponse, error) {
	account, err := s.uc.BanUser(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.BanUserResponse{User: userItem(*account)}, nil
}

func (s *IdentityService) PermitUser(ctx context.Context, req *v1.PermitUserRequest) (*v1.PermitUserResponse, error) {
	account, err := s.uc.PermitUser(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.PermitUserResponse{User: userItem(*account)}, nil
}

func (s *IdentityService) CreateRole(ctx context.Context, req *v1.CreateRoleRequest) (*v1.CreateRoleResponse, error) {
	item, err := s.uc.CreateRole(ctx, req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.CreateRoleResponse{Role: roleItem(*item)}, nil
}

func (s *IdentityService) UpdateRole(ctx context.Context, req *v1.UpdateRoleRequest) (*v1.UpdateRoleResponse, error) {
	item, err := s.uc.UpdateRole(ctx, req.GetUid(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRoleResponse{Role: roleItem(*item)}, nil
}

func (s *IdentityService) UpdateRoleStatus(ctx context.Context, req *v1.UpdateRoleStatusRequest) (*v1.UpdateRoleStatusResponse, error) {
	item, err := s.uc.UpdateRoleStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRoleStatusResponse{Role: roleItem(*item)}, nil
}

func (s *IdentityService) DeleteRole(ctx context.Context, req *v1.DeleteRoleRequest) (*v1.DeleteRoleResponse, error) {
	if err := s.uc.DeleteRole(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteRoleResponse{}, nil
}

func (s *IdentityService) GetRole(ctx context.Context, req *v1.GetRoleRequest) (*v1.GetRoleResponse, error) {
	item, err := s.uc.GetRole(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetRoleResponse{Role: roleItem(*item)}, nil
}

func (s *IdentityService) ListRoles(ctx context.Context, req *v1.ListRolesRequest) (*v1.ListRolesResponse, error) {
	items, total, page, pageSize, err := s.uc.ListRoles(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListRolesResponse{Items: roleItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *IdentityService) SelectRoles(ctx context.Context, req *v1.SelectRolesRequest) (*v1.SelectRolesResponse, error) {
	items, err := s.uc.SelectRoles(ctx, req.GetLimit(), req.GetLastUid())
	if err != nil {
		return nil, err
	}
	return &v1.SelectRolesResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *IdentityService) UpdateRolePermissions(ctx context.Context, req *v1.UpdateRolePermissionsRequest) (*v1.UpdateRolePermissionsResponse, error) {
	item, err := s.uc.UpdateRolePermissions(ctx, req.GetUid(), req.GetPermissions())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRolePermissionsResponse{Role: roleItem(*item)}, nil
}

func (s *IdentityService) ListPermissions(ctx context.Context, _ *v1.ListPermissionsRequest) (*v1.ListPermissionsResponse, error) {
	items, err := s.uc.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.PermissionItem, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.PermissionItem{Uid: item.UID, Name: item.Name, Resource: item.Resource, Action: item.Action})
	}
	return &v1.ListPermissionsResponse{Items: out}, nil
}

func tokenPair(session *biz.Session) *v1.TokenPair {
	return &v1.TokenPair{AccessToken: session.AccessToken, RefreshToken: session.RefreshToken, User: userItem(session.User)}
}

func userItem(account biz.Account) *v1.UserItem {
	return &v1.UserItem{
		Uid: account.UID, Name: account.Name, Nickname: account.Nickname,
		Email: account.Email, Avatar: account.Avatar, Remark: account.Remark, Status: account.Status,
	}
}

func userItems(items []biz.Account) []*v1.UserItem {
	out := make([]*v1.UserItem, 0, len(items))
	for _, item := range items {
		out = append(out, userItem(item))
	}
	return out
}

func namespaceItem(item biz.Namespace) *v1.NamespaceItem {
	return &v1.NamespaceItem{
		Uid: item.UID, Name: item.Name, OwnerUid: item.OwnerUID,
		Logo: item.Logo, Remark: item.Remark, Status: item.Status,
	}
}

func namespaceItems(items []biz.Namespace) []*v1.NamespaceItem {
	out := make([]*v1.NamespaceItem, 0, len(items))
	for _, item := range items {
		out = append(out, namespaceItem(item))
	}
	return out
}

func memberItem(item biz.Member) *v1.MemberItem {
	return &v1.MemberItem{
		Uid: item.UID, NamespaceUid: item.NamespaceUID, UserUid: item.UserUID,
		Email: item.Email, Name: item.Name, Status: item.Status,
		RoleUids: item.RoleUIDs, RoleNames: item.RoleNames,
	}
}

func memberItems(items []biz.Member) []*v1.MemberItem {
	out := make([]*v1.MemberItem, 0, len(items))
	for _, item := range items {
		out = append(out, memberItem(item))
	}
	return out
}

func roleItem(item biz.Role) *v1.RoleItem {
	return &v1.RoleItem{
		Uid: item.UID, NamespaceUid: item.NamespaceUID, Name: item.Name,
		Status: item.Status, IsSystem: item.IsSystem, Permissions: item.Permissions,
	}
}

func roleItems(items []biz.Role) []*v1.RoleItem {
	out := make([]*v1.RoleItem, 0, len(items))
	for _, item := range items {
		out = append(out, roleItem(item))
	}
	return out
}

func selectItems(items []biz.SelectItem) []*v1.SelectItem {
	out := make([]*v1.SelectItem, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.SelectItem{Uid: item.UID, Name: item.Name})
	}
	return out
}

func lastUID(items []biz.SelectItem) string {
	if len(items) == 0 {
		return ""
	}
	return items[len(items)-1].UID
}
