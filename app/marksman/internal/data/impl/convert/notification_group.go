package convert

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func toNotificationMemberDo(m *bo.NotificationMemberBo) *do.NotificationMember {
	if m == nil {
		return nil
	}
	return &do.NotificationMember{
		MemberUID: m.MemberUID,
		IsEmail:   m.IsEmail,
		IsPhone:   m.IsPhone,
	}
}

func toNotificationMemberBo(m *do.NotificationMember) *bo.NotificationMemberBo {
	if m == nil {
		return nil
	}
	return &bo.NotificationMemberBo{
		MemberUID: m.MemberUID,
		IsEmail:   m.IsEmail,
		IsPhone:   m.IsPhone,
	}
}

func ToNotificationGroupItemBo(m *do.NotificationGroup) *bo.NotificationGroupItemBo {
	if m == nil {
		return nil
	}
	var members []*bo.NotificationMemberBo
	if m.Members != nil {
		for _, mm := range m.Members.List() {
			members = append(members, toNotificationMemberBo(mm))
		}
	}
	var webhooks, templates []int64
	if m.Webhooks != nil {
		webhooks = m.Webhooks.List()
	}
	if m.Templates != nil {
		templates = m.Templates.List()
	}
	var emailConfigs []int64
	if m.EmailConfigs != nil {
		emailConfigs = m.EmailConfigs.List()
	}
	var metadata map[string]string
	if m.Metadata != nil {
		metadata = m.Metadata.Map()
	}
	return &bo.NotificationGroupItemBo{
		UID:          m.ID,
		Name:         m.Name,
		Remark:       m.Remark,
		Metadata:     metadata,
		Status:       m.Status,
		Members:      members,
		Webhooks:     webhooks,
		Templates:    templates,
		EmailConfigs: emailConfigs,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func ToNotificationGroupDo(ctx context.Context, req *bo.CreateNotificationGroupBo) *do.NotificationGroup {
	if req == nil {
		return nil
	}
	members := make([]*do.NotificationMember, 0, len(req.Members))
	for _, m := range req.Members {
		members = append(members, toNotificationMemberDo(m))
	}
	m := &do.NotificationGroup{
		Name:      req.Name,
		Remark:    req.Remark,
		Metadata:  safety.NewMap(req.Metadata),
		Status:    enum.GlobalStatus_ENABLED,
		Members:   safety.NewSlice(members),
		Webhooks:  safety.NewSlice(req.Webhooks),
		Templates: safety.NewSlice(req.Templates),
		EmailConfigs: safety.NewSlice(req.EmailConfigs),
	}
	m.WithNamespace(contextx.GetNamespace(ctx)).WithCreator(contextx.GetUserUID(ctx))
	return m
}

func ToNotificationGroupDoUpdate(req *bo.UpdateNotificationGroupBo) *do.NotificationGroup {
	if req == nil {
		return nil
	}
	members := make([]*do.NotificationMember, 0, len(req.Members))
	for _, m := range req.Members {
		members = append(members, toNotificationMemberDo(m))
	}
	return &do.NotificationGroup{
		Name:      req.Name,
		Remark:    req.Remark,
		Metadata:  safety.NewMap(req.Metadata),
		Members:   safety.NewSlice(members),
		Webhooks:  safety.NewSlice(req.Webhooks),
		Templates: safety.NewSlice(req.Templates),
		EmailConfigs: safety.NewSlice(req.EmailConfigs),
	}
}
