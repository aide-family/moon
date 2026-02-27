package bo

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/magicbox/api/v1"
)

type UserItemBo struct {
	UID       snowflake.ID
	Email     string
	Name      string
	Nickname  string
	Avatar    string
	Status    enum.UserStatus
	Remark    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (b *UserItemBo) ToAPIV1UserItem() *apiv1.UserItem {
	return &apiv1.UserItem{
		Uid:       b.UID.Int64(),
		Email:     b.Email,
		Name:      b.Name,
		Nickname:  b.Nickname,
		Avatar:    b.Avatar,
		Status:    b.Status,
		Remark:    b.Remark,
		CreatedAt: b.CreatedAt.Format(time.DateTime),
		UpdatedAt: b.UpdatedAt.Format(time.DateTime),
	}
}

type ListUserBo struct {
	*PageRequestBo
	Email   string
	Keyword string
	Status  enum.UserStatus
}

func NewListUserBo(req *apiv1.ListUserRequest) *ListUserBo {
	return &ListUserBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		Email:         req.Email,
		Keyword:       req.Keyword,
		Status:        req.Status,
	}
}

func ToAPIV1ListUserReply(pageResponseBo *PageResponseBo[*UserItemBo]) *apiv1.ListUserReply {
	items := make([]*apiv1.UserItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, item.ToAPIV1UserItem())
	}
	return &apiv1.ListUserReply{
		Items:    items,
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
	}
}

type SelectUserBo struct {
	Keyword string
	Limit   int32
	LastUID snowflake.ID
	Status  enum.UserStatus
}

func NewSelectUserBo(req *apiv1.SelectUserRequest) *SelectUserBo {
	var lastUID snowflake.ID
	if req.LastUID > 0 {
		lastUID = snowflake.ParseInt64(req.LastUID)
	}
	return &SelectUserBo{
		Keyword: req.Keyword,
		Limit:   req.Limit,
		LastUID: lastUID,
		Status:  req.Status,
	}
}

type SelectUserItemBo struct {
	Value    snowflake.ID
	Label    string
	Disabled bool
	Tooltip  string
}

type SelectUserBoResult struct {
	Items   []*SelectUserItemBo
	Total   int64
	LastUID snowflake.ID
	HasMore bool
}

func ToAPIV1SelectUserReply(result *SelectUserBoResult) *apiv1.SelectUserReply {
	items := make([]*apiv1.SelectUserItem, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, &apiv1.SelectUserItem{
			Value:    item.Value.Int64(),
			Label:    item.Label,
			Disabled: item.Disabled,
			Tooltip:  item.Tooltip,
		})
	}
	return &apiv1.SelectUserReply{
		Items:   items,
		Total:   result.Total,
		LastUID: result.LastUID.Int64(),
		HasMore: result.HasMore,
	}
}
