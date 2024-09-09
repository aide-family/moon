package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	dictapi "github.com/aide-family/moon/api/admin/dict"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ IDictModuleBuilder = (*dictModuleBuilder)(nil)

type (
	dictModuleBuilder struct {
		ctx context.Context
	}

	IDictModuleBuilder interface {
		WithCreateDictRequest(*dictapi.CreateDictRequest) ICreateDictRequestBuilder
		WithUpdateDictRequest(*dictapi.UpdateDictRequest) IUpdateDictRequestBuilder
		WithListDictRequest(*dictapi.ListDictRequest) IListDictRequestBuilder
		WithUpdateDictStatusParams(*dictapi.BatchUpdateDictStatusRequest) IUpdateDictStatusParamsBuilder
		DoDictBuilder() IDoDictBuilder
		DictTypeList() []*api.EnumItem
	}

	ICreateDictRequestBuilder interface {
		ToBo() *bo.CreateDictParams
	}

	createDictRequestBuilder struct {
		ctx context.Context
		*dictapi.CreateDictRequest
	}

	IUpdateDictRequestBuilder interface {
		ToBo() *bo.UpdateDictParams
	}

	updateDictRequestBuilder struct {
		ctx context.Context
		*dictapi.UpdateDictRequest
	}

	IListDictRequestBuilder interface {
		ToBo() *bo.QueryDictListParams
	}

	listDictRequestBuilder struct {
		ctx context.Context
		*dictapi.ListDictRequest
	}

	IUpdateDictStatusParamsBuilder interface {
		ToBo() *bo.UpdateDictStatusParams
	}

	updateDictStatusParamsBuilder struct {
		ctx context.Context
		*dictapi.BatchUpdateDictStatusRequest
	}

	IDoDictBuilder interface {
		ToAPI(imodel.IDict) *adminapi.DictItem
		ToAPIs([]imodel.IDict) []*adminapi.DictItem
		ToSelect(imodel.IDict) *adminapi.SelectItem
		ToSelects([]imodel.IDict) []*adminapi.SelectItem
	}

	doDictBuilder struct {
		ctx context.Context
	}
)

func (d *doDictBuilder) ToAPI(dict imodel.IDict) *adminapi.DictItem {
	if types.IsNil(d) || types.IsNil(dict) {
		return nil
	}

	return &adminapi.DictItem{
		Id:           dict.GetID(),
		Name:         dict.GetName(),
		DictType:     api.DictType(dict.GetDictType()),
		ColorType:    dict.GetColorType(),
		CssClass:     dict.GetCSSClass(),
		Value:        dict.GetValue(),
		Icon:         dict.GetIcon(),
		ImageUrl:     dict.GetImageURL(),
		Status:       api.Status(dict.GetStatus()),
		LanguageCode: dict.GetLanguageCode().String(),
		Remark:       dict.GetRemark(),
		CreatedAt:    dict.GetCreatedAt().String(),
		UpdatedAt:    dict.GetUpdatedAt().String(),
	}
}

func (d *doDictBuilder) ToAPIs(dicts []imodel.IDict) []*adminapi.DictItem {
	if types.IsNil(d) || types.IsNil(dicts) {
		return nil
	}

	return types.SliceTo(dicts, func(item imodel.IDict) *adminapi.DictItem {
		return d.ToAPI(item)
	})
}

func (d *doDictBuilder) ToSelect(dict imodel.IDict) *adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(dict) {
		return nil
	}

	return &adminapi.SelectItem{
		Value:    dict.GetID(),
		Label:    dict.GetName(),
		Children: nil,
		Disabled: dict.GetDeletedAt() > 0 || !dict.GetStatus().IsEnable(),
		Extend: &adminapi.SelectExtend{
			Icon:   dict.GetIcon(),
			Color:  dict.GetCSSClass(),
			Remark: dict.GetRemark(),
			Image:  dict.GetImageURL(),
		},
	}
}

func (d *doDictBuilder) ToSelects(dicts []imodel.IDict) []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(dicts) {
		return nil
	}

	return types.SliceTo(dicts, func(item imodel.IDict) *adminapi.SelectItem {
		return d.ToSelect(item)
	})
}

func (u *updateDictStatusParamsBuilder) ToBo() *bo.UpdateDictStatusParams {
	if types.IsNil(u) || types.IsNil(u.BatchUpdateDictStatusRequest) {
		return nil
	}

	return &bo.UpdateDictStatusParams{
		IDs:    u.GetIds(),
		Status: vobj.Status(u.Status),
	}
}

func (l *listDictRequestBuilder) ToBo() *bo.QueryDictListParams {
	if types.IsNil(l) || types.IsNil(l.ListDictRequest) {
		return nil
	}

	return &bo.QueryDictListParams{
		Keyword:  l.Keyword,
		Page:     types.NewPagination(l.GetPagination()),
		Status:   vobj.Status(l.Status),
		DictType: vobj.DictType(l.GetDictType()),
	}
}

func (u *updateDictRequestBuilder) ToBo() *bo.UpdateDictParams {
	if types.IsNil(u) || types.IsNil(u.UpdateDictRequest) {
		return nil
	}

	return &bo.UpdateDictParams{
		ID:          u.GetId(),
		UpdateParam: NewParamsBuild().WithContext(u.ctx).DictModuleBuilder().WithCreateDictRequest(u.GetData()).ToBo(),
	}
}

func (c *createDictRequestBuilder) ToBo() *bo.CreateDictParams {
	if types.IsNil(c) || types.IsNil(c.CreateDictRequest) {
		return nil
	}

	return &bo.CreateDictParams{
		Name:         c.GetName(),
		Remark:       c.GetRemark(),
		Value:        c.GetValue(),
		DictType:     vobj.DictType(c.GetDictType()),
		ColorType:    c.GetColorType(),
		CSSClass:     c.GetCssClass(),
		Icon:         c.GetIcon(),
		ImageURL:     c.GetImageUrl(),
		Status:       vobj.Status(c.GetStatus()),
		LanguageCode: vobj.ToLanguage(c.GetLanguageCode()),
	}
}

func (d *dictModuleBuilder) WithCreateDictRequest(request *dictapi.CreateDictRequest) ICreateDictRequestBuilder {
	return &createDictRequestBuilder{ctx: d.ctx, CreateDictRequest: request}
}

func (d *dictModuleBuilder) WithUpdateDictRequest(request *dictapi.UpdateDictRequest) IUpdateDictRequestBuilder {
	return &updateDictRequestBuilder{ctx: d.ctx, UpdateDictRequest: request}
}

func (d *dictModuleBuilder) WithListDictRequest(request *dictapi.ListDictRequest) IListDictRequestBuilder {
	return &listDictRequestBuilder{ctx: d.ctx, ListDictRequest: request}
}

func (d *dictModuleBuilder) WithUpdateDictStatusParams(request *dictapi.BatchUpdateDictStatusRequest) IUpdateDictStatusParamsBuilder {
	return &updateDictStatusParamsBuilder{ctx: d.ctx, BatchUpdateDictStatusRequest: request}
}

func (d *dictModuleBuilder) DoDictBuilder() IDoDictBuilder {
	return &doDictBuilder{ctx: d.ctx}
}

func (d *dictModuleBuilder) DictTypeList() []*api.EnumItem {
	dictTypeList := []vobj.DictType{
		vobj.DictTypeStrategyCategory,
		vobj.DictTypeStrategyGroupCategory,
		vobj.DictTypeAlarmLevel,
		vobj.DictTypeAlarmPage,
	}
	return types.SliceTo(dictTypeList, func(item vobj.DictType) *api.EnumItem {
		return &api.EnumItem{
			Value: int32(item),
			Label: item.String(),
		}
	})
}
