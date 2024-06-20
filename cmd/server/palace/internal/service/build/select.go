package build

import (
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	types2 "github.com/aide-family/moon/pkg/util/types"
)

type SelectBuild struct {
	*bo.SelectOptionBo
}

func NewSelectBuild(option *bo.SelectOptionBo) *SelectBuild {
	return &SelectBuild{
		SelectOptionBo: option,
	}
}

func (b *SelectBuild) ToApi() *admin.Select {
	if types2.IsNil(b) || types2.IsNil(b.SelectOptionBo) {
		return nil
	}
	return &admin.Select{
		Value: b.Value,
		Label: b.Label,
		Children: types2.SliceTo(b.Children, func(i *bo.SelectOptionBo) *admin.Select {
			return NewSelectBuild(i).ToApi()
		}),
		Disabled: b.Disabled,
		Extend:   SelectExtendToApi(b.Extend),
	}
}

func SelectExtendToApi(extend *bo.SelectExtend) *admin.SelectExtend {
	if types2.IsNil(extend) {
		return nil
	}
	return &admin.SelectExtend{
		Icon:   extend.Icon,
		Color:  extend.Color,
		Remark: extend.Remark,
		Image:  extend.Image,
	}
}
