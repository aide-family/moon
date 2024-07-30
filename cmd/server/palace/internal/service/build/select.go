package build

import (
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"
)

// SelectBuilder 下拉选择器构造器
type SelectBuilder struct {
	*bo.SelectOptionBo
}

// NewSelectBuilder 创建下拉选择器构造器
func NewSelectBuilder(option *bo.SelectOptionBo) *SelectBuilder {
	return &SelectBuilder{
		SelectOptionBo: option,
	}
}

// ToAPI 转换为API对象
func (b *SelectBuilder) ToAPI() *admin.SelectItem {
	if types.IsNil(b) || types.IsNil(b.SelectOptionBo) {
		return nil
	}
	return &admin.SelectItem{
		Value: b.Value,
		Label: b.Label,
		Children: types.SliceTo(b.Children, func(i *bo.SelectOptionBo) *admin.SelectItem {
			return NewSelectBuilder(i).ToAPI()
		}),
		Disabled: b.Disabled,
		Extend:   SelectExtendToAPI(b.Extend),
	}
}

// SelectExtendToAPI 转换为API对象
func SelectExtendToAPI(extend *bo.SelectExtend) *admin.SelectExtend {
	if types.IsNil(extend) {
		return nil
	}
	return &admin.SelectExtend{
		Icon:   extend.Icon,
		Color:  extend.Color,
		Remark: extend.Remark,
		Image:  extend.Image,
	}
}
