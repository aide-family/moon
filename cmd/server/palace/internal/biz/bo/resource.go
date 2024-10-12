package bo

import (
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// QueryResourceListParams 查询资源列表请求参数
	QueryResourceListParams struct {
		Keyword string           `json:"keyword"`
		Page    types.Pagination `json:"page"`
		IsAll   bool             `json:"isAll"`
		Status  vobj.Status      `json:"status"`
	}

	// ResourceSelectOptionBuild 资源选项构建器
	ResourceSelectOptionBuild struct {
		imodel.IResource
	}

	// QueryTeamMenuListParams 查询团队菜单列表请求参数
	QueryTeamMenuListParams struct {
		TeamID uint32 `json:"teamID"`
	}
)

// NewResourceSelectOptionBuild 构建资源选项构建器
func NewResourceSelectOptionBuild(resource imodel.IResource) *ResourceSelectOptionBuild {
	return &ResourceSelectOptionBuild{
		IResource: resource,
	}
}

// ToSelectOption 转换为选项
func (b *ResourceSelectOptionBuild) ToSelectOption() *SelectOptionBo {
	if types.IsNil(b) || types.IsNil(b.IResource) {
		return nil
	}
	return &SelectOptionBo{
		Value:    b.GetID(),
		Label:    b.GetName(),
		Disabled: !b.GetStatus().IsEnable(),
	}
}
