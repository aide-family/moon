package bo

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

type MemberItemBo struct {
	UID       snowflake.ID
	Email     string
	Status    enum.MemberStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	UserUID   snowflake.ID
	Name      string
	Nickname  string
	Avatar    string
	Remark    string
}

func (b *MemberItemBo) ToAPIV1MemberItem() *goddessv1.MemberItem {
	return &goddessv1.MemberItem{
		Uid:       b.UID.Int64(),
		Email:     b.Email,
		Status:    b.Status,
		CreatedAt: b.CreatedAt.Format(time.DateTime),
		UpdatedAt: b.UpdatedAt.Format(time.DateTime),
		UserUID:   b.UserUID.Int64(),
		Name:      b.Name,
		Nickname:  b.Nickname,
		Avatar:    b.Avatar,
		Remark:    b.Remark,
	}
}

type InviteMemberBo struct {
	Email   string
	RoleUID uint32
}

type CreateMemberBo struct {
	Creator      snowflake.ID
	NamespaceUID snowflake.ID
	UserUID      snowflake.ID
	Name         string
	Nickname     string
	Avatar       string
	Status       enum.MemberStatus
}

type ListMemberBo struct {
	*PageRequestBo
	UserUID snowflake.ID
	Status  enum.MemberStatus
	UIDs    []int64
	Keyword string
	Email   string
	Phone   string
}

func NewListMemberBo(req *goddessv1.ListMemberRequest) *ListMemberBo {
	return &ListMemberBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		UserUID:       snowflake.ParseInt64(req.UserUID),
		Status:        req.Status,
		UIDs:          req.Uids,
		Keyword:       req.Keyword,
		Email:         req.Email,
		Phone:         req.Phone,
	}
}

func ToAPIV1ListMemberReply(pageResponseBo *PageResponseBo[*MemberItemBo]) *goddessv1.ListMemberReply {
	items := make([]*goddessv1.MemberItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, item.ToAPIV1MemberItem())
	}
	return &goddessv1.ListMemberReply{
		Items:    items,
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
	}
}

type SelectMemberBo struct {
	Keyword string
	Limit   int32
	LastUID snowflake.ID
	Status  enum.MemberStatus
}

func NewSelectMemberBo(req *goddessv1.SelectMemberRequest) *SelectMemberBo {
	var lastUID snowflake.ID
	if req.LastUID > 0 {
		lastUID = snowflake.ParseInt64(req.LastUID)
	}
	return &SelectMemberBo{
		Keyword: req.Keyword,
		Limit:   req.Limit,
		LastUID: lastUID,
		Status:  req.Status,
	}
}

type SelectMemberItemBo struct {
	Value    snowflake.ID
	Label    string
	Disabled bool
	Tooltip  string
}

type SelectMemberBoResult struct {
	Items   []*SelectMemberItemBo
	Total   int64
	LastUID snowflake.ID
	HasMore bool
}

func ToAPIV1SelectMemberReply(result *SelectMemberBoResult) *goddessv1.SelectMemberReply {
	items := make([]*goddessv1.SelectMemberItem, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, &goddessv1.SelectMemberItem{
			Value:    item.Value.Int64(),
			Label:    item.Label,
			Disabled: item.Disabled,
			Tooltip:  item.Tooltip,
		})
	}
	return &goddessv1.SelectMemberReply{
		Items:   items,
		Total:   result.Total,
		LastUID: result.LastUID.Int64(),
		HasMore: result.HasMore,
	}
}
