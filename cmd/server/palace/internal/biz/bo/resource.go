package bo

import (
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/model"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
)

type (
	QueryResourceListParams struct {
		Keyword string `json:"keyword"`
		Page    types.Pagination
	}

	ResourceSelectOptionBuild struct {
		*model.SysAPI
	}
)

// NewResourceSelectOptionBuild 构建资源选项构建器
func NewResourceSelectOptionBuild(resource *model.SysAPI) *ResourceSelectOptionBuild {
	return &ResourceSelectOptionBuild{
		SysAPI: resource,
	}
}

// ToSelectOption 转换为选项
func (b *ResourceSelectOptionBuild) ToSelectOption() *SelectOptionBo {
	if types.IsNil(b) || types.IsNil(b.SysAPI) {
		return nil
	}
	return &SelectOptionBo{
		Value:    b.ID,
		Label:    b.Name,
		Disabled: !vobj.Status(b.Status).IsEnable(),
	}
}
