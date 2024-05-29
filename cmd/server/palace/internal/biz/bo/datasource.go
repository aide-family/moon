package bo

import (
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
)

type (
	CreateDatasourceParams struct {
		// 数据源名称
		Name string `json:"name"`
		// 数据源类型
		Type vobj.DatasourceType `json:"type"`
		// 数据源地址
		Endpoint string `json:"endpoint"`
		// 状态
		Status vobj.Status `json:"status"`
		// 描述
		Remark string `json:"remark"`
		// 数据源配置(json 字符串)
		Config string `json:"config"`
	}

	QueryDatasourceListParams struct {
		// 分页, 不传不分页
		Page types.Pagination `json:"page"`
		// 关键字
		Keyword string `json:"keyword"`
		// 团队ID
		TeamID uint32 `json:"teamID"`
		// 数据源类型
		Type vobj.DatasourceType `json:"type"`
		// 状态
		Status vobj.Status `json:"status"`
	}

	UpdateDatasourceBaseInfoParams struct {
		ID uint32 `json:"id"`
		// 数据源名称
		Name string `json:"name"`
		// 状态
		Status vobj.Status `json:"status"`
		// 描述
		Remark string `json:"remark"`
	}

	UpdateDatasourceConfigParams struct {
		ID uint32 `json:"id"`
		// 数据源配置(json 字符串)
		Config string `json:"config"`
		// 数据源类型
		Type vobj.DatasourceType `json:"type"`
	}

	DatasourceOptionBuild struct {
		*bizmodel.Datasource
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
