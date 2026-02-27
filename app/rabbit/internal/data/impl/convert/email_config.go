// Package convert provides conversion functions for the application.
package convert

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/strutil"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/data/impl/do"
)

func ToEmailConfigDo(ctx context.Context, req *bo.CreateEmailConfigBo) *do.EmailConfig {
	model := &do.EmailConfig{
		NamespaceUID: contextx.GetNamespace(ctx),
		Name:         req.Name,
		Host:         req.Host,
		Port:         req.Port,
		Username:     req.Username,
		Password:     strutil.EncryptString(req.Password),
		Status:       enum.GlobalStatus_ENABLED,
	}
	model.WithCreator(contextx.GetUserUID(ctx))
	return model
}

func ToEmailConfigBO(emailConfigDO *do.EmailConfig) *bo.EmailConfigItemBo {
	return &bo.EmailConfigItemBo{
		UID:       emailConfigDO.ID,
		Name:      emailConfigDO.Name,
		Host:      emailConfigDO.Host,
		Port:      emailConfigDO.Port,
		Username:  emailConfigDO.Username,
		Password:  string(emailConfigDO.Password),
		Status:    emailConfigDO.Status,
		CreatedAt: emailConfigDO.CreatedAt,
		UpdatedAt: emailConfigDO.UpdatedAt,
	}
}

func ToEmailConfigItemSelectBO(emailConfigDO *do.EmailConfig) *bo.EmailConfigItemSelectBo {
	return &bo.EmailConfigItemSelectBo{
		UID:      emailConfigDO.ID,
		Name:     emailConfigDO.Name,
		Status:   enum.GlobalStatus(emailConfigDO.Status),
		Disabled: emailConfigDO.Status != enum.GlobalStatus_ENABLED || emailConfigDO.DeletedAt.Valid,
		Tooltip:  emailConfigDO.Name,
	}
}
