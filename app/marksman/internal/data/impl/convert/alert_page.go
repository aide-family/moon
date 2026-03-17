package convert

import (
	"context"

	"github.com/aide-family/magicbox/contextx"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func ToAlertPageItemBo(m *do.AlertPage) *bo.AlertPageItemBo {
	if m == nil {
		return nil
	}
	var filter *bo.AlertPageFilterBo
	if m.FilterConfig != nil {
		filter = &bo.AlertPageFilterBo{
			StrategyGroupUIDs: m.FilterConfig.StrategyGroupUIDs,
			LevelUIDs:         m.FilterConfig.LevelUIDs,
			StrategyUIDs:      m.FilterConfig.StrategyUIDs,
		}
	}
	return &bo.AlertPageItemBo{
		UID:       m.ID,
		Name:      m.Name,
		Color:     m.Color,
		SortOrder: m.SortOrder,
		Filter:    filter,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func ToAlertPageDo(ctx context.Context, req *bo.CreateAlertPageBo) *do.AlertPage {
	if req == nil {
		return nil
	}
	m := &do.AlertPage{
		Name:      req.Name,
		Color:     req.Color,
		SortOrder: req.SortOrder,
	}
	if filter := req.Filter; filter != nil {
		m.FilterConfig = &do.AlertPageFilterConfig{
			StrategyGroupUIDs: filter.StrategyGroupUIDs,
			LevelUIDs:         filter.LevelUIDs,
			StrategyUIDs:      filter.StrategyUIDs,
		}
	}
	m.WithNamespace(contextx.GetNamespace(ctx)).WithCreator(contextx.GetUserUID(ctx))
	return m
}

func ToAlertPageDoUpdate(req *bo.UpdateAlertPageBo) (*do.AlertPage, *do.AlertPageFilterConfig) {
	if req == nil {
		return nil, nil
	}
	m := &do.AlertPage{
		Name:      req.Name,
		Color:     req.Color,
		SortOrder: req.SortOrder,
	}
	var filter *do.AlertPageFilterConfig
	if filterBo := req.Filter; filterBo != nil {
		filter = &do.AlertPageFilterConfig{
			StrategyGroupUIDs: filterBo.StrategyGroupUIDs,
			LevelUIDs:         filterBo.LevelUIDs,
			StrategyUIDs:      filterBo.StrategyUIDs,
		}
	}
	return m, filter
}
