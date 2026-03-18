package bo

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

type NotificationMemberBo struct {
	MemberUID int64
	IsEmail   bool
	IsPhone   bool
}

func NewNotificationMemberBo(req *apiv1.NotificationMemberItem) *NotificationMemberBo {
	if req == nil {
		return nil
	}
	return &NotificationMemberBo{
		MemberUID: req.GetMemberUid(),
		IsEmail:   req.GetIsEmail(),
		IsPhone:   req.GetIsPhone(),
	}
}

func NewNotificationMembersBo(req []*apiv1.NotificationMemberItem) []*NotificationMemberBo {
	if req == nil {
		return nil
	}
	members := make([]*NotificationMemberBo, 0, len(req))
	for _, m := range req {
		members = append(members, NewNotificationMemberBo(m))
	}
	return members
}

type NotificationGroupItemBo struct {
	UID       snowflake.ID
	Name      string
	Remark    string
	Metadata  map[string]string
	Status    enum.GlobalStatus
	Members   []*NotificationMemberBo
	Webhooks  []int64
	Templates []int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ToAPIV1NotificationMemberItem(b *NotificationMemberBo) *apiv1.NotificationMemberItem {
	if b == nil {
		return nil
	}
	return &apiv1.NotificationMemberItem{
		MemberUid: b.MemberUID,
		IsEmail:   b.IsEmail,
		IsPhone:   b.IsPhone,
	}
}

func ToAPIV1NotificationGroupItem(b *NotificationGroupItemBo) *apiv1.NotificationGroupItem {
	if b == nil {
		return nil
	}
	members := make([]*apiv1.NotificationMemberItem, 0, len(b.Members))
	for _, m := range b.Members {
		members = append(members, ToAPIV1NotificationMemberItem(m))
	}
	return &apiv1.NotificationGroupItem{
		Uid:       b.UID.Int64(),
		Name:      b.Name,
		Remark:    b.Remark,
		Metadata:  b.Metadata,
		Status:    b.Status,
		Members:   members,
		Webhooks:  b.Webhooks,
		Templates: b.Templates,
		CreatedAt: timex.FormatTime(&b.CreatedAt),
		UpdatedAt: timex.FormatTime(&b.UpdatedAt),
	}
}

type CreateNotificationGroupBo struct {
	Name      string
	Remark    string
	Metadata  map[string]string
	Members   []*NotificationMemberBo
	Webhooks  []int64
	Templates []int64
}

func NewCreateNotificationGroupBo(req *apiv1.CreateNotificationGroupRequest) *CreateNotificationGroupBo {
	if req == nil {
		return nil
	}
	members := NewNotificationMembersBo(req.GetMembers())
	return &CreateNotificationGroupBo{
		Name:      req.GetName(),
		Remark:    req.GetRemark(),
		Metadata:  req.GetMetadata(),
		Members:   members,
		Webhooks:  req.GetWebhooks(),
		Templates: req.GetTemplates(),
	}
}

type UpdateNotificationGroupBo struct {
	UID       snowflake.ID
	Name      string
	Remark    string
	Metadata  map[string]string
	Members   []*NotificationMemberBo
	Webhooks  []int64
	Templates []int64
}

func NewUpdateNotificationGroupBo(req *apiv1.UpdateNotificationGroupRequest) *UpdateNotificationGroupBo {
	if req == nil {
		return nil
	}
	members := NewNotificationMembersBo(req.GetMembers())
	return &UpdateNotificationGroupBo{
		UID:       snowflake.ParseInt64(req.GetUid()),
		Name:      req.GetName(),
		Remark:    req.GetRemark(),
		Metadata:  req.GetMetadata(),
		Members:   members,
		Webhooks:  req.GetWebhooks(),
		Templates: req.GetTemplates(),
	}
}

type UpdateNotificationGroupStatusBo struct {
	UID    snowflake.ID
	Status enum.GlobalStatus
}

func NewUpdateNotificationGroupStatusBo(req *apiv1.UpdateNotificationGroupStatusRequest) *UpdateNotificationGroupStatusBo {
	if req == nil {
		return nil
	}
	return &UpdateNotificationGroupStatusBo{
		UID:    snowflake.ParseInt64(req.GetUid()),
		Status: req.GetStatus(),
	}
}

type ListNotificationGroupBo struct {
	*PageRequestBo
	Keyword string
	Status  enum.GlobalStatus
}

func NewListNotificationGroupBo(req *apiv1.ListNotificationGroupRequest) *ListNotificationGroupBo {
	if req == nil {
		return nil
	}
	return &ListNotificationGroupBo{
		PageRequestBo: NewPageRequestBo(req.GetPage(), req.GetPageSize()),
		Keyword:       req.GetKeyword(),
		Status:        req.GetStatus(),
	}
}

func ToAPIV1ListNotificationGroupReply(page *PageResponseBo[*NotificationGroupItemBo]) *apiv1.ListNotificationGroupReply {
	if page == nil {
		return nil
	}
	items := make([]*apiv1.NotificationGroupItem, 0, len(page.GetItems()))
	for _, item := range page.GetItems() {
		items = append(items, ToAPIV1NotificationGroupItem(item))
	}
	return &apiv1.ListNotificationGroupReply{
		Items:    items,
		Total:    page.GetTotal(),
		Page:     page.GetPage(),
		PageSize: page.GetPageSize(),
	}
}
