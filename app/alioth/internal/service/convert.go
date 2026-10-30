package service

import (
	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

func userItem(account bo.Account) *v1.UserItem {
	return &v1.UserItem{
		Uid: account.UID, Name: account.Name, Nickname: account.Nickname,
		Email: account.Email, Phone: account.Phone, Avatar: account.Avatar, Remark: account.Remark, Status: account.Status,
		SystemRole: account.SystemRole, AuthorizedBy: account.AuthorizedBy,
	}
}

func userItems(items []bo.Account) []*v1.UserItem {
	out := make([]*v1.UserItem, 0, len(items))
	for _, item := range items {
		out = append(out, userItem(item))
	}
	return out
}

func namespaceItem(item bo.Namespace) *v1.NamespaceItem {
	createdAt := int64(0)
	if !item.CreatedAt.IsZero() {
		createdAt = item.CreatedAt.Unix()
	}
	// 套餐到期时间为零表示永久，对外统一输出 0。
	packageExpiresAt := int64(0)
	if !item.PackageExpiresAt.IsZero() {
		packageExpiresAt = item.PackageExpiresAt.Unix()
	}
	return &v1.NamespaceItem{
		Uid: item.UID, Name: item.Name, OwnerUid: item.OwnerUID,
		OwnerName: item.OwnerName, OwnerEmail: item.OwnerEmail,
		Logo: item.Logo, Remark: item.Remark, Status: item.Status,
		MemberCount: int32(item.MemberCount), CreatedAt: createdAt,
		PackageName: item.PackageName, PackageExpiresAt: packageExpiresAt,
		DisableReason: item.DisableReason, MaxMembers: item.MaxMembers,
	}
}

func namespaceItems(items []bo.Namespace) []*v1.NamespaceItem {
	out := make([]*v1.NamespaceItem, 0, len(items))
	for _, item := range items {
		out = append(out, namespaceItem(item))
	}
	return out
}

func memberItem(item bo.Member) *v1.MemberItem {
	return &v1.MemberItem{
		Uid: item.UID, NamespaceUid: item.NamespaceUID, UserUid: item.UserUID,
		Email: item.Email, Phone: item.Phone, Name: item.Name, Status: item.Status,
		RoleUids: item.RoleUIDs, RoleNames: item.RoleNames,
	}
}

func memberItems(items []bo.Member) []*v1.MemberItem {
	out := make([]*v1.MemberItem, 0, len(items))
	for _, item := range items {
		out = append(out, memberItem(item))
	}
	return out
}

func roleItem(item bo.Role) *v1.RoleItem {
	return &v1.RoleItem{
		Uid: item.UID, NamespaceUid: item.NamespaceUID, Name: item.Name,
		Status: item.Status, IsSystem: item.IsSystem, Remark: item.Remark, Permissions: item.Permissions,
	}
}

func roleItems(items []bo.Role) []*v1.RoleItem {
	out := make([]*v1.RoleItem, 0, len(items))
	for _, item := range items {
		out = append(out, roleItem(item))
	}
	return out
}

func inviteItem(item bo.MemberInvite) *v1.MemberInviteItem {
	return &v1.MemberInviteItem{
		Uid: item.UID, NamespaceUid: item.NamespaceUID, NamespaceName: item.NamespaceName,
		Email: item.Email, RoleUid: item.RoleUID, RoleName: item.RoleName,
		InviterName: item.InviterName, Status: item.Status,
		ExpiresAt: item.ExpiresAt.Unix(), CreatedAt: item.CreatedAt.Unix(),
		Registered: item.UserUID != "",
	}
}

func inviteItems(items []bo.MemberInvite) []*v1.MemberInviteItem {
	out := make([]*v1.MemberInviteItem, 0, len(items))
	for _, item := range items {
		out = append(out, inviteItem(item))
	}
	return out
}

func systemMessageItem(item bo.SystemMessage) *v1.SystemMessageItem {
	return &v1.SystemMessageItem{
		Uid: item.UID, Kind: item.Kind, ActorName: item.ActorName,
		NamespaceUid: item.NamespaceUID, NamespaceName: item.NamespaceName,
		RoleName: item.RoleName, Read: item.ReadAt != nil,
		CreatedAt: item.CreatedAt.Unix(),
	}
}

func systemMessageItems(items []bo.SystemMessage) []*v1.SystemMessageItem {
	out := make([]*v1.SystemMessageItem, 0, len(items))
	for _, item := range items {
		out = append(out, systemMessageItem(item))
	}
	return out
}

func selectItems(items []bo.SelectItem) []*v1.SelectItem {
	out := make([]*v1.SelectItem, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.SelectItem{Uid: item.UID, Name: item.Name})
	}
	return out
}

func lastUID(items []bo.SelectItem) string {
	if len(items) == 0 {
		return ""
	}
	return items[len(items)-1].UID
}

func featureItem(item bo.Feature) *v1.FeatureItem {
	return &v1.FeatureItem{
		Code: item.Code, Module: item.Module, Name: item.Name,
		Description: item.Description, Permissions: item.Permissions,
	}
}

func featureItems(items []bo.Feature) []*v1.FeatureItem {
	out := make([]*v1.FeatureItem, 0, len(items))
	for _, item := range items {
		out = append(out, featureItem(item))
	}
	return out
}

func featurePackageItem(item bo.FeaturePackage) *v1.FeaturePackageItem {
	createdAt := int64(0)
	if !item.CreatedAt.IsZero() {
		createdAt = item.CreatedAt.Unix()
	}
	return &v1.FeaturePackageItem{
		Uid: item.UID, Name: item.Name, Remark: item.Remark, Features: item.Features,
		IsBase: item.IsBase, Status: item.Status, CreatedAt: createdAt,
		DurationDays: item.DurationDays, Visibility: item.Visibility,
		MaxMembers: item.MaxMembers, PriceCents: item.PriceCents,
	}
}

// namespaceOrderItem 把订单快照转成接口视图。时间字段零值统一输出 0。
func namespaceOrderItem(item bo.NamespaceOrder) *v1.NamespaceOrderItem {
	createdAt := int64(0)
	if !item.CreatedAt.IsZero() {
		createdAt = item.CreatedAt.Unix()
	}
	paidAt := int64(0)
	if !item.PaidAt.IsZero() {
		paidAt = item.PaidAt.Unix()
	}
	return &v1.NamespaceOrderItem{
		Uid: item.UID, NamespaceUid: item.NamespaceUID,
		NamespaceName: item.NamespaceName, NamespaceRemark: item.NamespaceRemark,
		PackageUid: item.PackageUID, PackageName: item.PackageName,
		PriceCents: item.PriceCents, DurationDays: item.DurationDays,
		MaxMembers: item.MaxMembers, Status: item.Status,
		CreatedAt: createdAt, PaidAt: paidAt,
	}
}

func namespaceOrderItems(items []bo.NamespaceOrder) []*v1.NamespaceOrderItem {
	out := make([]*v1.NamespaceOrderItem, 0, len(items))
	for _, item := range items {
		out = append(out, namespaceOrderItem(item))
	}
	return out
}

func featurePackageItems(items []bo.FeaturePackage) []*v1.FeaturePackageItem {
	out := make([]*v1.FeaturePackageItem, 0, len(items))
	for _, item := range items {
		out = append(out, featurePackageItem(item))
	}
	return out
}

// namespaceFeatureFields 拆出租户能力开通的响应字段，GET 和 PUT 共用同一份视图。
// 到期时间为零表示永久，对外统一输出 0。
func namespaceFeatureFields(item bo.NamespaceFeature) (packageUID, packageName string, features, packageFeatures, extraFeatures []string, expiresAt int64, expired bool, durationDays int32) {
	if !item.ExpiresAt.IsZero() {
		expiresAt = item.ExpiresAt.Unix()
	}
	return item.PackageUID, item.PackageName, item.Features, item.PackageFeatures,
		item.ExtraFeatures, expiresAt, item.Expired, item.DurationDays
}
