package convert

import (
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func ToDatasourceItemBo(m *do.Datasource) *bo.DatasourceItemBo {
	if m == nil {
		return nil
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
	}
}

func ToDatasourceDo(req *bo.CreateDatasourceBo) *do.Datasource {
	if req == nil {
		return nil
	}
	return &do.Datasource{
		Name:     req.Name,
		Type:     req.Type,
		Driver:   req.Driver,
		Metadata: safety.NewMap(req.Metadata),
		Status:   enum.GlobalStatus_ENABLED,
	}
}
