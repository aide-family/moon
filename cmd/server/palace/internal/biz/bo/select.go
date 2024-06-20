package bo

import (
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

type (
	SelectExtend struct {
		Icon, Color, Remark, Image string
	}
	SelectOptionBo struct {
		Value    uint32            `json:"value"`
		Label    string            `json:"label"`
		Disabled bool              `json:"disabled"`
		Children []*SelectOptionBo `json:"children"`
		Extend   *SelectExtend     `json:"extend"`
	}

	DatasourceOptionBuild struct {
		*bizmodel.Datasource
	}

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

func NewDatasourceMetricOptionBuild(metric *bizmodel.DatasourceMetric) *DatasourceMetricOptionBuild {
	return &DatasourceMetricOptionBuild{
		DatasourceMetric: metric,
	}
}

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
