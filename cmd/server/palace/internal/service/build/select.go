package build

import (
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"
)

type SelectBuilder struct {
	*bo.SelectOptionBo
}

func NewSelectBuilder(option *bo.SelectOptionBo) *SelectBuilder {
	return &SelectBuilder{
		SelectOptionBo: option,
	}
}

func (b *SelectBuilder) ToApi() *admin.Select {
	if types.IsNil(b) || types.IsNil(b.SelectOptionBo) {
		return nil
	}
	return &admin.Select{
		Value: b.Value,
		Label: b.Label,
		Children: types.SliceTo(b.Children, func(i *bo.SelectOptionBo) *admin.Select {
			return NewSelectBuilder(i).ToApi()
		}),
		Disabled: b.Disabled,
		Extend:   SelectExtendToApi(b.Extend),
	}
}

func SelectExtendToApi(extend *bo.SelectExtend) *admin.SelectExtend {
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
