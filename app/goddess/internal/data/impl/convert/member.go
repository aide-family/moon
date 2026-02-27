package convert

import (
	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/do"
)

func MemberToBo(member *do.Member, email string) *bo.MemberItemBo {
	if member == nil {
		return nil
	}
	return &bo.MemberItemBo{
		UID:       member.UID,
		Email:     email,
		Status:    member.Status,
		CreatedAt: member.CreatedAt,
		UpdatedAt: member.UpdatedAt,
		UserUID:   member.UserUID,
		Name:      member.Name,
		Nickname:  member.Nickname,
		Avatar:    member.Avatar,
		Remark:    member.Remark,
	}
}
