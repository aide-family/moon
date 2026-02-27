package convert

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/strutil"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/data/impl/do"
)

func ToRecipientMemberDo(ctx context.Context, req *bo.RecipientMemberItemBo) *do.RecipientMember {
	m := &do.RecipientMember{
		BaseModel: do.BaseModel{
			ID: req.UID,
		},
		NamespaceUID: contextx.GetNamespace(ctx),
		UserUID:      req.UserUID,
		Email:        strutil.EncryptString(req.Email),
		Phone:        strutil.EncryptString(req.Phone),
		Status:       req.Status,
		Name:         req.Name,
		Nickname:     req.Nickname,
		Avatar:       req.Avatar,
		Remark:       req.Remark,
	}
	m.WithCreator(contextx.GetUserUID(ctx))
	return m
}

func ToRecipientMemberBO(d *do.RecipientMember) *bo.RecipientMemberItemBo {
	return &bo.RecipientMemberItemBo{
		UID:          d.ID,
		NamespaceUID: d.NamespaceUID,
		UserUID:      d.UserUID,
		Email:        string(d.Email),
		Phone:        string(d.Phone),
		Status:       d.Status,
		Name:         d.Name,
		Nickname:     d.Nickname,
		Avatar:       d.Avatar,
		Remark:       d.Remark,
	}
}
