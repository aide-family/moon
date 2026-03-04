// Package bo is the business logic object
package bo

import (
	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

type RecipientMemberItemBo struct {
	UID      snowflake.ID      `json:"uid"`
	UserUID  snowflake.ID      `json:"user_uid"`
	Name     string            `json:"name"`
	Nickname string            `json:"nickname"`
	Avatar   string            `json:"avatar"`
	Remark   string            `json:"remark"`
	Email    string            `json:"email"`
	Phone    string            `json:"phone"`
	Status   enum.MemberStatus `json:"status"`
}

func (b *RecipientMemberItemBo) ToAPIV1MemberItem() *goddessv1.MemberItem {
	return &goddessv1.MemberItem{
		Uid:      b.UID.Int64(),
		UserUID:  b.UserUID.Int64(),
		Status:   b.Status,
		Name:     b.Name,
		Nickname: b.Nickname,
		Avatar:   b.Avatar,
		Remark:   b.Remark,
		Email:    b.Email,
		Phone:    b.Phone,
	}
}

func NewRecipientMemberItemBo(members []*goddessv1.MemberItem) []*RecipientMemberItemBo {
	out := make([]*RecipientMemberItemBo, 0, len(members))
	for _, m := range members {
		out = append(out, &RecipientMemberItemBo{
			UID:      snowflake.ID(m.Uid),
			UserUID:  snowflake.ID(m.UserUID),
			Status:   m.Status,
			Name:     m.Name,
			Nickname: m.Nickname,
			Avatar:   m.Avatar,
			Remark:   m.Remark,
			Email:    m.Email,
			Phone:    m.Phone,
		})
	}
	return out
}
