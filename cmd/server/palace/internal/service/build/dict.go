package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	dictapi "github.com/aide-family/moon/api/admin/dict"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/imodel"
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
		SysDict imodel.IDict

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
		Id:           b.SysDict.GetID(),
		Name:         b.SysDict.GetName(),
		Value:        b.SysDict.GetValue(),
		ColorType:    b.SysDict.GetColorType(),
		Icon:         b.SysDict.GetIcon(),
		Status:       api.Status(b.SysDict.GetStatus()),
		DictType:     api.DictType(b.SysDict.GetDictType()),
		ImageUrl:     b.SysDict.GetImageUrl(),
		LanguageCode: b.SysDict.GetLanguageCode(),
		Remark:       b.SysDict.GetRemark(),
		CreatedAt:    b.SysDict.GetCreatedAt().String(),
		UpdatedAt:    b.SysDict.GetUpdatedAt().String(),
	}
}

// ToApiSelect 转换成api下拉数据
func (b *dictBuilder) ToApiSelect() *admin.Select {
	if types.IsNil(b) || types.IsNil(b.SysDict) {
		return nil
	}
	return &admin.Select{
		Value:    b.SysDict.GetID(),
		Label:    b.SysDict.GetName(),
		Children: nil,
		Disabled: !b.SysDict.GetStatus().IsEnable() || b.SysDict.GetDeletedAt() > 0,
		Extend: &admin.SelectExtend{
			Icon:   b.SysDict.GetIcon(),
			Color:  b.SysDict.GetCssClass(),
			Remark: b.SysDict.GetRemark(),
			Image:  b.SysDict.GetImageUrl(),
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
