package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
)

type DictBuild struct {
	*model.SysDict
}

func NewDictBuild(dict *model.SysDict) *DictBuild {
	return &DictBuild{
		SysDict: dict,
	}
}

// ToApi 转换成api
func (b *DictBuild) ToApi() *admin.Dict {
	if types.IsNil(b) || types.IsNil(b.SysDict) {
		return nil
	}
	return &admin.Dict{
		Id:           b.ID,
		Name:         b.Name,
		Value:        b.Value,
		ColorType:    b.ColorType,
		Icon:         b.Icon,
		Status:       api.Status(b.Status),
		DictType:     api.DictType(b.DictType),
		ImageUrl:     b.ImageUrl,
		LanguageCode: b.LanguageCode,
		Remark:       b.Remark,
		CreatedAt:    b.CreatedAt.String(),
		UpdatedAt:    b.UpdatedAt.String(),
	}
}
