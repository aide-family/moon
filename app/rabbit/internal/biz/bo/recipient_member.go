// Package bo is the business logic object
package bo

import (
	magicboxapiv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"
)

type RecipientMemberItemBo struct {
	UID          snowflake.ID      `json:"uid"`
	NamespaceUID snowflake.ID      `json:"namespace_uid"`
	UserUID      snowflake.ID      `json:"user_uid"`
	Name         string            `json:"name"`
	Nickname     string            `json:"nickname"`
	Avatar       string            `json:"avatar"`
	Remark       string            `json:"remark"`
	Email        string            `json:"email"`
	Phone        string            `json:"phone"`
	Status       enum.MemberStatus `json:"status"`
}

func (b *RecipientMemberItemBo) ToAPIV1MemberItem() *magicboxapiv1.MemberItem {
	return &magicboxapiv1.MemberItem{
		Uid:          b.UID.Int64(),
		NamespaceUID: b.NamespaceUID.Int64(),
		UserUID:      b.UserUID.Int64(),
		Status:       b.Status,
		Name:         b.Name,
		Nickname:     b.Nickname,
		Avatar:       b.Avatar,
		Remark:       b.Remark,
		Email:        b.Email,
		Phone:        b.Phone,
	}
}

func NewRecipientMemberItemBo(members []*magicboxapiv1.MemberItem) []*RecipientMemberItemBo {
	out := make([]*RecipientMemberItemBo, 0, len(members))
	for _, m := range members {
		out = append(out, &RecipientMemberItemBo{
			UID:          snowflake.ID(m.Uid),
			NamespaceUID: snowflake.ID(m.NamespaceUID),
			UserUID:      snowflake.ID(m.UserUID),
			Status:       m.Status,
			Name:         m.Name,
			Nickname:     m.Nickname,
			Avatar:       m.Avatar,
			Remark:       m.Remark,
			Email:        m.Email,
			Phone:        m.Phone,
		})
	}
	return out
}
