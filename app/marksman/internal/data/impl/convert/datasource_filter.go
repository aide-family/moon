package convert

import (
	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func ToDatasourceFilterBo(filter *do.DatasourceFilter) *bo.DatasourceFilterBo {
	if filter == nil {
		return &bo.DatasourceFilterBo{}
	}
	return &bo.DatasourceFilterBo{
		DatasourceUIDs:          filter.DatasourceUIDs,
		ExcludeDatasourceUIDs:   filter.ExcludeDatasourceUIDs,
		DatasourceLabels:        filter.DatasourceLabels,
		ExcludeDatasourceLabels: filter.ExcludeDatasourceLabels,
	}
}

func ToDatasourceFilterDo(filter *bo.DatasourceFilterBo) *do.DatasourceFilter {
	if filter == nil {
		return &do.DatasourceFilter{}
	}
	return &do.DatasourceFilter{
		DatasourceUIDs:          filter.DatasourceUIDs,
		ExcludeDatasourceUIDs:   filter.ExcludeDatasourceUIDs,
		DatasourceLabels:        filter.DatasourceLabels,
		ExcludeDatasourceLabels: filter.ExcludeDatasourceLabels,
	}
}
