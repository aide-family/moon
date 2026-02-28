package convert

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/safety"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func ToStrategyItemBo(m *do.Strategy) *bo.StrategyItemBo {
	if m == nil {
		return nil
	}
	return &bo.StrategyItemBo{
		UID:              m.ID,
		Name:             m.Name,
		Remark:           m.Remark,
		Type:             m.Type,
		Driver:           m.Driver,
		Status:           m.Status,
		Metadata:         m.Metadata.Map(),
		StrategyGroupUID: m.StrategyGroupUID,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func ToStrategyDo(ctx context.Context, req *bo.CreateStrategyBo) *do.Strategy {
	if req == nil {
		return nil
	}
	model := &do.Strategy{
		Name:             req.Name,
		Remark:           req.Remark,
		Type:             req.Type,
		Driver:           req.Driver,
		StrategyGroupUID: req.StrategyGroupUID,
		Metadata:         safety.NewMap(req.Metadata),
		Status:           req.Status,
	}
	model.WithNamespace(contextx.GetNamespace(ctx)).WithCreator(contextx.GetUserUID(ctx))
	return model
}
