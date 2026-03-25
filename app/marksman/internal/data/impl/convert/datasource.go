package convert

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func ToDatasourceItemBo(m *do.Datasource) *bo.DatasourceItemBo {
	if m == nil {
		return nil
	}
	levelName := ""
	if m.Level != nil {
		levelName = m.Level.Name
	}
	return &bo.DatasourceItemBo{
		UID:       m.ID,
		Name:      m.Name,
		Type:      m.Type,
		Driver:    m.Driver,
		Metadata:  m.Metadata.Map(),
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		URL:       m.URL,
		Remark:    m.Remark,
		LevelUID:  m.LevelUID,
		LevelName: levelName,
	}
}

func ToDatasourceDo(ctx context.Context, req *bo.CreateDatasourceBo) *do.Datasource {
	if req == nil {
		return nil
	}
	model := &do.Datasource{
		Name:     req.Name,
		Type:     req.Type,
		Driver:   req.Driver,
		Metadata: safety.NewMap(req.Metadata),
		Status:   enum.GlobalStatus_ENABLED,
		URL:      req.URL,
		Remark:   req.Remark,
		LevelUID: req.LevelUID,
	}
	model.WithNamespace(contextx.GetNamespace(ctx)).WithCreator(contextx.GetUserUID(ctx))
	return model
}

func ToSelectDatasourceItemBo(m *do.Datasource) *bo.SelectDatasourceItemBo {
	if m == nil {
		return nil
	}
	return &bo.SelectDatasourceItemBo{
		Value:    m.ID,
		Label:    m.Name,
		Disabled: m.Status != enum.GlobalStatus_ENABLED || m.DeletedAt.Valid,
		Tooltip:  m.Remark,
		Type:     m.Type,
		Driver:   m.Driver,
		URL:      m.URL,
	}
}
