package build

import (
	"github.com/aide-cloud/moon/api/admin"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/pkg/types"
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
	if types.IsNil(b) || types.IsNil(b.SelectOptionBo) {
		return nil
	}
	return &admin.Select{
		Value: b.Value,
		Label: b.Label,
		Children: types.SliceTo(b.Children, func(i *bo.SelectOptionBo) *admin.Select {
			return NewSelectBuild(i).ToApi()
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
