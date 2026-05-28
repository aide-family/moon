package bo

import apiv1 "github.com/aide-family/rabbit/pkg/api/v1"

// NotificationMemberBo stores a namespace member with per-channel delivery preferences.
type NotificationMemberBo struct {
	MemberUID    int64
	IsEmail      bool
	IsSMS        bool
	IsPhone      bool
	MemberName   string
	MemberAvatar string
	MemberEmail  string
	MemberPhone  string
}

// AlertSubscriptionMemberBo is an alias for alert subscription member preferences.
type AlertSubscriptionMemberBo = NotificationMemberBo

// RecipientGroupMemberBo is an alias for recipient group member preferences.
type RecipientGroupMemberBo = NotificationMemberBo

func NewAlertSubscriptionMembersBo(reqs []*apiv1.AlertSubscriptionMemberRequest) []*NotificationMemberBo {
	return newNotificationMembersBoFromAlertRequests(reqs)
}

func NewRecipientGroupMembersBo(reqs []*apiv1.RecipientGroupMemberRequest) []*NotificationMemberBo {
	items := make([]*NotificationMemberBo, 0, len(reqs))
	for _, item := range reqs {
		if item == nil {
			continue
		}
		items = append(items, &NotificationMemberBo{
			MemberUID: item.GetMemberUid(),
			IsEmail:   item.GetIsEmail(),
			IsSMS:     item.GetIsSms(),
			IsPhone:   item.GetIsPhone(),
		})
	}
	return items
}

func newNotificationMembersBoFromAlertRequests(reqs []*apiv1.AlertSubscriptionMemberRequest) []*NotificationMemberBo {
	items := make([]*NotificationMemberBo, 0, len(reqs))
	for _, item := range reqs {
		if item == nil {
			continue
		}
		items = append(items, &NotificationMemberBo{
			MemberUID: item.GetMemberUid(),
			IsEmail:   item.GetIsEmail(),
			IsSMS:     item.GetIsSms(),
			IsPhone:   item.GetIsPhone(),
		})
	}
	return items
}

func (b *NotificationMemberBo) ToAPIV1AlertSubscriptionMember() *apiv1.AlertSubscriptionMemberItem {
	if b == nil {
		return nil
	}
	return &apiv1.AlertSubscriptionMemberItem{
		MemberUid:    b.MemberUID,
		IsEmail:      b.IsEmail,
		IsSms:        b.IsSMS,
		IsPhone:      b.IsPhone,
		MemberName:   b.MemberName,
		MemberAvatar: b.MemberAvatar,
		MemberEmail:  b.MemberEmail,
		MemberPhone:  b.MemberPhone,
	}
}

func (b *NotificationMemberBo) ToAPIV1RecipientGroupMember() *apiv1.RecipientGroupMemberItem {
	if b == nil {
		return nil
	}
	return &apiv1.RecipientGroupMemberItem{
		MemberUid:    b.MemberUID,
		IsEmail:      b.IsEmail,
		IsSms:        b.IsSMS,
		IsPhone:      b.IsPhone,
		MemberName:   b.MemberName,
		MemberAvatar: b.MemberAvatar,
		MemberEmail:  b.MemberEmail,
		MemberPhone:  b.MemberPhone,
	}
}
