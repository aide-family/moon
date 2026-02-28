package convert

import (
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func ToStrategyGroupItemBo(m *do.StrategyGroup) *bo.StrategyGroupItemBo {
	if m == nil {
		return nil
	}
	return &bo.StrategyGroupItemBo{
		UID:       m.ID,
		Name:      m.Name,
		Remark:    m.Remark,
		Status:    m.Status,
		Metadata:  m.Metadata.Map(),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func ToStrategyGroupItemSelectBo(m *do.StrategyGroup) *bo.StrategyGroupItemSelectBo {
	if m == nil {
		return nil
	}
	return &bo.StrategyGroupItemSelectBo{
		Value:    m.ID.Int64(),
		Label:    m.Name,
		Disabled: m.Status != enum.GlobalStatus_ENABLED || m.DeletedAt.Valid,
		Tooltip:  m.Remark,
	}
}

func ToStrategyGroupDo(req *bo.CreateStrategyGroupBo) *do.StrategyGroup {
	if req == nil {
		return nil
	}
	return &do.StrategyGroup{
		Name:     req.Name,
		Remark:   req.Remark,
		Metadata: safety.NewMap(req.Metadata),
		Status:   enum.GlobalStatus_ENABLED,
	}
}
