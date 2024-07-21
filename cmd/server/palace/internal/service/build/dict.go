package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	dictapi "github.com/aide-family/moon/api/admin/dict"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	DictModelBuilder interface {
		ToApi() *admin.Dict
		ToApiSelect() *admin.Select
	}

	DictRequestBuilder interface {
		ToCreateDictBO() *bo.CreateDictParams

		ToUpdateDictBO() *bo.UpdateDictParams
	}

	dictBuilder struct {
		// model
		*model.SysDict

		// request
		CreateDictRequest *dictapi.CreateDictRequest
		UpdateDictRequest *dictapi.UpdateDictRequest

		// context
		ctx context.Context
	}
)

// ToApi 转换成api
func (b *dictBuilder) ToApi() *admin.Dict {
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

// ToApiSelect 转换成api下拉数据
func (b *dictBuilder) ToApiSelect() *admin.Select {
	if types.IsNil(b) || types.IsNil(b.SysDict) {
		return nil
	}
	return &admin.Select{
		Value:    b.ID,
		Label:    b.Name,
		Children: nil,
		Disabled: !b.Status.IsEnable() || b.DeletedAt > 0,
		Extend: &admin.SelectExtend{
			Icon:   b.Icon,
			Color:  b.CssClass,
			Remark: b.Remark,
			Image:  b.ImageUrl,
		},
	}
}

func (b *dictBuilder) ToCreateDictBO() *bo.CreateDictParams {
	return &bo.CreateDictParams{
		Name:         b.CreateDictRequest.GetName(),
		Value:        b.CreateDictRequest.GetValue(),
		DictType:     vobj.DictType(b.CreateDictRequest.GetDictType()),
		ColorType:    b.CreateDictRequest.GetColorType(),
		CssClass:     b.CreateDictRequest.GetCssClass(),
		Icon:         b.CreateDictRequest.GetIcon(),
		ImageUrl:     b.CreateDictRequest.GetImageUrl(),
		Status:       vobj.Status(b.CreateDictRequest.GetStatus()),
		Remark:       b.CreateDictRequest.GetRemark(),
		LanguageCode: b.CreateDictRequest.GetLanguageCode(),
	}
}

func (b *dictBuilder) ToUpdateDictBO() *bo.UpdateDictParams {
	data := b.UpdateDictRequest.GetData()
	createParams := bo.CreateDictParams{
		Name:         data.GetName(),
		Value:        data.GetValue(),
		DictType:     vobj.DictType(data.GetDictType()),
		ColorType:    data.GetColorType(),
		CssClass:     data.GetCssClass(),
		Icon:         data.GetIcon(),
		ImageUrl:     data.GetImageUrl(),
		Status:       vobj.Status(data.GetStatus()),
		Remark:       data.GetRemark(),
		LanguageCode: data.GetLanguageCode(),
	}
	return &bo.UpdateDictParams{
		ID:          b.UpdateDictRequest.GetId(),
		UpdateParam: createParams,
	}
}

type DictTypeBuilder struct {
	list []vobj.DictType
}

func NewDictTypeBuilder() *DictTypeBuilder {
	return &DictTypeBuilder{
		list: []vobj.DictType{
			vobj.DictTypePromLabel,
			vobj.DictTypePromAnnotation,
			vobj.DictTypeNotifyType,
			vobj.DictTypePromStrategyGroup,
			vobj.DictTypeAlarmLevel,
			vobj.DictTypeAlarmPage,
			vobj.DictTypeAlarmStatus,
			vobj.DictTypePromStrategy,
			vobj.DictTypeStrategyCategory,
		},
	}
}

func (b *DictTypeBuilder) ToApi() []*api.EnumItem {
	if types.IsNil(b) {
		return nil
	}
	var list []*api.EnumItem
	for _, item := range b.list {
		list = append(list, &api.EnumItem{
			Value: int32(item),
			Label: item.String(),
		})
	}
	return list
}
