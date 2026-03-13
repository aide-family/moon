// Package bo is the business logic object
package bo

import (
	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

// RecipientGroupItemBo is the BO for one recipient group item.
type RecipientGroupItemBo struct {
	UID            snowflake.ID             `json:"uid"`
	Name           string                   `json:"name"`
	Metadata       map[string]string        `json:"metadata"`
	Status         enum.GlobalStatus        `json:"status"`
	Members        []*RecipientMemberItemBo `json:"members"`
	Templates      []*TemplateItemBo        `json:"templates"`
	EmailConfigs   []*EmailConfigItemBo     `json:"emailConfigs"`
	WebhookConfigs []*WebhookItemBo         `json:"webhookConfigs"`
}

// ToAPIV1RecipientGroupItem converts BO to API item.
func (b *RecipientGroupItemBo) ToAPIV1RecipientGroupItem() *apiv1.RecipientGroupItem {
	members := make([]*goddessv1.MemberItem, 0, len(b.Members))
	for _, m := range b.Members {
		members = append(members, m.ToAPIV1MemberItem())
	}
	templates := make([]*apiv1.TemplateItem, 0, len(b.Templates))
	for _, t := range b.Templates {
		templates = append(templates, t.ToAPIV1TemplateItem())
	}
	emailConfigs := make([]*apiv1.EmailConfigItem, 0, len(b.EmailConfigs))
	for _, e := range b.EmailConfigs {
		emailConfigs = append(emailConfigs, e.ToAPIV1EmailConfigItem())
	}
	webhookConfigs := make([]*apiv1.WebhookItem, 0, len(b.WebhookConfigs))
	for _, w := range b.WebhookConfigs {
		webhookConfigs = append(webhookConfigs, w.ToAPIV1WebhookItem())
	}
	return &apiv1.RecipientGroupItem{
		Uid:            b.UID.Int64(),
		Name:           b.Name,
		Metadata:       b.Metadata,
		Members:        members,
		Templates:      templates,
		EmailConfigs:   emailConfigs,
		WebhookConfigs: webhookConfigs,
		Status:         b.Status,
	}
}

// CreateRecipientGroupBo is the BO for creating a recipient group.
type CreateRecipientGroupBo struct {
	Name           string
	Metadata       map[string]string
	Templates      []int64
	EmailConfigs   []int64
	WebhookConfigs []int64
	Members        []int64
}

// NewCreateRecipientGroupBo builds BO from API request.
func NewCreateRecipientGroupBo(req *apiv1.CreateRecipientGroupRequest) *CreateRecipientGroupBo {
	return &CreateRecipientGroupBo{
		Name:           req.Name,
		Metadata:       req.Metadata,
		Templates:      req.Templates,
		EmailConfigs:   req.EmailConfigs,
		WebhookConfigs: req.WebhookConfigs,
		Members:        req.Members,
	}
}

// UpdateRecipientGroupBo is the BO for updating a recipient group.
type UpdateRecipientGroupBo struct {
	UID            snowflake.ID
	Name           string
	Metadata       map[string]string
	Templates      []int64
	EmailConfigs   []int64
	WebhookConfigs []int64
	Members        []int64
}

// NewUpdateRecipientGroupBo builds BO from API request.
func NewUpdateRecipientGroupBo(req *apiv1.UpdateRecipientGroupRequest) *UpdateRecipientGroupBo {
	bo := &UpdateRecipientGroupBo{
		UID:            snowflake.ParseInt64(req.Uid),
		Name:           req.Name,
		Metadata:       req.Metadata,
		Templates:      req.Templates,
		EmailConfigs:   req.EmailConfigs,
		WebhookConfigs: req.WebhookConfigs,
		Members:        req.Members,
	}
	return bo
}

// ListRecipientGroupBo is the BO for list recipient group request.
type ListRecipientGroupBo struct {
	*PageRequestBo
	Keyword string
	Status  enum.GlobalStatus
}

// NewListRecipientGroupBo builds BO from API request.
func NewListRecipientGroupBo(req *apiv1.ListRecipientGroupRequest) *ListRecipientGroupBo {
	return &ListRecipientGroupBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		Keyword:       req.Keyword,
		Status:        req.Status,
	}
}

// ToAPIV1ListRecipientGroupReply converts BO to API list reply.
func ToAPIV1ListRecipientGroupReply(page *PageResponseBo[*RecipientGroupItemBo]) *apiv1.ListRecipientGroupReply {
	items := make([]*apiv1.RecipientGroupItem, 0, len(page.GetItems()))
	for _, it := range page.GetItems() {
		items = append(items, it.ToAPIV1RecipientGroupItem())
	}
	return &apiv1.ListRecipientGroupReply{
		Items:    items,
		Total:    page.GetTotal(),
		Page:     page.GetPage(),
		PageSize: page.GetPageSize(),
	}
}

// SelectRecipientGroupBo is the BO for select recipient group request.
type SelectRecipientGroupBo struct {
	Keyword string
	Limit   int32
	LastUID snowflake.ID
	Status  enum.GlobalStatus
}

// NewSelectRecipientGroupBo builds BO from API request.
func NewSelectRecipientGroupBo(req *apiv1.SelectRecipientGroupRequest) *SelectRecipientGroupBo {
	var lastUID snowflake.ID
	if req.LastUID > 0 {
		lastUID = snowflake.ParseInt64(req.LastUID)
	}
	return &SelectRecipientGroupBo{
		Keyword: req.GetKeyword(),
		Limit:   req.GetLimit(),
		LastUID: lastUID,
		Status:  req.GetStatus(),
	}
}

// SelectRecipientGroupItemBo is one item in select recipient group result.
type SelectRecipientGroupItemBo struct {
	UID      snowflake.ID
	Name     string
	Status   enum.GlobalStatus
	Disabled bool
	Tooltip  string
}

// ToAPIV1SelectRecipientGroupItem converts BO to API item.
func (b *SelectRecipientGroupItemBo) ToAPIV1SelectRecipientGroupItem() *apiv1.SelectRecipientGroupItem {
	return &apiv1.SelectRecipientGroupItem{
		Value:    b.UID.Int64(),
		Label:    b.Name,
		Disabled: b.Disabled,
		Tooltip:  b.Tooltip,
	}
}

// SelectRecipientGroupBoResult is the result of select recipient group.
type SelectRecipientGroupBoResult struct {
	Items   []*SelectRecipientGroupItemBo
	Total   int64
	LastUID snowflake.ID
}

// ToAPIV1SelectRecipientGroupReply converts BO result to API reply.
func ToAPIV1SelectRecipientGroupReply(items []*SelectRecipientGroupItemBo, total int64, lastUID snowflake.ID, limit int32) *apiv1.SelectRecipientGroupReply {
	out := make([]*apiv1.SelectRecipientGroupItem, 0, len(items))
	for _, it := range items {
		out = append(out, it.ToAPIV1SelectRecipientGroupItem())
	}
	hasMore := int32(len(items)) == limit
	return &apiv1.SelectRecipientGroupReply{
		Items:   out,
		Total:   total,
		LastUID: lastUID.Int64(),
		HasMore: hasMore,
	}
}

// UpdateRecipientGroupStatusBo is the BO for updating recipient group status.
type UpdateRecipientGroupStatusBo struct {
	UID    snowflake.ID
	Status enum.GlobalStatus
}

// NewUpdateRecipientGroupStatusBo builds BO from API request.
func NewUpdateRecipientGroupStatusBo(req *apiv1.UpdateRecipientGroupStatusRequest) *UpdateRecipientGroupStatusBo {
	return &UpdateRecipientGroupStatusBo{
		UID:    snowflake.ParseInt64(req.Uid),
		Status: req.Status,
	}
}

// RecipientGroupDetailBo is the detail BO (includes related IDs for edit form).
type RecipientGroupDetailBo struct {
	RecipientGroupItemBo
	Templates      []snowflake.ID
	EmailConfigs   []snowflake.ID
	WebhookConfigs []snowflake.ID
	Members        []snowflake.ID
}

// ToAPIV1RecipientGroupItemFromDetail converts detail BO to API item (proto has uid/name/metadata; relations are sent on Update).
func (b *RecipientGroupDetailBo) ToAPIV1RecipientGroupItemFromDetail() *apiv1.RecipientGroupItem {
	return &apiv1.RecipientGroupItem{
		Uid:      b.UID.Int64(),
		Name:     b.Name,
		Metadata: b.Metadata,
	}
}
