package bo

import (
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

type (
	// SelectExtend 选择项扩展
	SelectExtend struct {
		Icon, Color, Remark, Image string
	}

	// SelectOptionBo 选择项明细
	SelectOptionBo struct {
		Value    uint32            `json:"value"`
		Label    string            `json:"label"`
		Disabled bool              `json:"disabled"`
		Children []*SelectOptionBo `json:"children"`
		Extend   *SelectExtend     `json:"extend"`
	}

	// DatasourceOptionBuild 数据源选项构建器
	DatasourceOptionBuild struct {
		*bizmodel.Datasource
	}

	// DatasourceMetricOptionBuild 数据源指标选项构建器
	DatasourceMetricOptionBuild struct {
		*bizmodel.DatasourceMetric
	}
)

// NewDatasourceOptionBuild 创建数据源选项构建器
func NewDatasourceOptionBuild(datasource *bizmodel.Datasource) *DatasourceOptionBuild {
	return &DatasourceOptionBuild{
		Datasource: datasource,
	}
}

// ToSelectOption 转换为选择项
func (b *DatasourceOptionBuild) ToSelectOption() *SelectOptionBo {
	return &SelectOptionBo{
		Value:    b.ID,
		Label:    b.Name,
		Disabled: b.DeletedAt > 0 || !b.Status.IsEnable(),
	}
}

// NewDatasourceMetricOptionBuild 创建数据源指标选项构建器
func NewDatasourceMetricOptionBuild(metric *bizmodel.DatasourceMetric) *DatasourceMetricOptionBuild {
	return &DatasourceMetricOptionBuild{
		DatasourceMetric: metric,
	}
}

// ToSelectOption 转换为选择项
func (b *DatasourceMetricOptionBuild) ToSelectOption() *SelectOptionBo {
	return &SelectOptionBo{
		Value:    b.ID,
		Label:    b.Name,
		Disabled: b.DeletedAt > 0,
		Children: nil,
		Extend: &SelectExtend{
			Icon:   "",
			Color:  "",
			Remark: b.Unit,
			Image:  "",
		},
	}
}
