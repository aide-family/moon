// Package convert provides conversion functions for level data.
package convert

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func ToLevelItemBo(m *do.Level) *bo.LevelItemBo {
	if m == nil {
		return nil
	}
	return &bo.LevelItemBo{
		UID:       m.ID,
		Name:      m.Name,
		Remark:    m.Remark,
		Status:    m.Status,
		Metadata:  m.Metadata.Map(),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func ToLevelItemSelectBo(m *do.Level) *bo.LevelItemSelectBo {
	if m == nil {
		return nil
	}
	return &bo.LevelItemSelectBo{
		Value:    m.ID.Int64(),
		Label:    m.Name,
		Disabled: m.Status != enum.GlobalStatus_ENABLED || m.DeletedAt.Valid,
		Tooltip:  m.Remark,
	}
}

func ToLevelDo(ctx context.Context, req *bo.CreateLevelBo) *do.Level {
	if req == nil {
		return nil
	}
	model := &do.Level{
		Name:     req.Name,
		Remark:   req.Remark,
		Metadata: safety.NewMap(req.Metadata),
		Status:   enum.GlobalStatus_ENABLED,
	}
	model.WithNamespace(contextx.GetNamespace(ctx)).WithCreator(contextx.GetUserUID(ctx))
	return model
}
