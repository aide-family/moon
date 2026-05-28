package convert

import (
	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/data/impl/do"
)

func ToNotificationMembersDO(items []*bo.NotificationMemberBo) do.NotificationMembers {
	out := make(do.NotificationMembers, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, do.NotificationMember{
			MemberUID: item.MemberUID,
			IsEmail:   item.IsEmail,
			IsSMS:     item.IsSMS,
			IsPhone:   item.IsPhone,
		})
	}
	return out
}

func ToNotificationMembersBo(items do.NotificationMembers) []*bo.NotificationMemberBo {
	out := make([]*bo.NotificationMemberBo, 0, len(items))
	for _, item := range items {
		out = append(out, &bo.NotificationMemberBo{
			MemberUID: item.MemberUID,
			IsEmail:   item.IsEmail,
			IsSMS:     item.IsSMS,
			IsPhone:   item.IsPhone,
		})
	}
	return out
}
