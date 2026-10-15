package service

import (
	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func userItem(account bo.Account) *v1.UserItem {
	return &v1.UserItem{
		Uid: account.UID, Name: account.Name, Nickname: account.Nickname,
		Email: account.Email, Avatar: account.Avatar, Remark: account.Remark, Status: account.Status,
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
	return &v1.NamespaceItem{
		Uid: item.UID, Name: item.Name, OwnerUid: item.OwnerUID,
		Logo: item.Logo, Remark: item.Remark, Status: item.Status,
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
		Email: item.Email, Name: item.Name, Status: item.Status,
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
		Status: item.Status, IsSystem: item.IsSystem, Permissions: item.Permissions,
	}
}

func roleItems(items []bo.Role) []*v1.RoleItem {
	out := make([]*v1.RoleItem, 0, len(items))
	for _, item := range items {
		out = append(out, roleItem(item))
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
