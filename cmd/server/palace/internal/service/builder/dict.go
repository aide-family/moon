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

	// IDictModuleBuilder 字典模块构造器
	IDictModuleBuilder interface {
		// WithCreateDictRequest 创建字典请求参数构造器
		WithCreateDictRequest(*dictapi.CreateDictRequest) ICreateDictRequestBuilder
		// WithUpdateDictRequest 更新字典请求参数构造器
		WithUpdateDictRequest(*dictapi.UpdateDictRequest) IUpdateDictRequestBuilder
		// WithListDictRequest 获取字典列表请求参数构造器
		WithListDictRequest(*dictapi.ListDictRequest) IListDictRequestBuilder
		// WithUpdateDictStatusParams 更新字典状态请求参数构造器
		WithUpdateDictStatusParams(*dictapi.BatchUpdateDictStatusRequest) IUpdateDictStatusParamsBuilder
		// DoDictBuilder 字典条目构造器
		DoDictBuilder() IDoDictBuilder
		// DictTypeList 字典类型列表
		DictTypeList() []*api.EnumItem
	}

	// ICreateDictRequestBuilder 创建字典请求参数构造器
	ICreateDictRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.CreateDictParams
	}

	createDictRequestBuilder struct {
		ctx context.Context
		*dictapi.CreateDictRequest
	}

	// IUpdateDictRequestBuilder 更新字典请求参数构造器
	IUpdateDictRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateDictParams
	}

	updateDictRequestBuilder struct {
		ctx context.Context
		*dictapi.UpdateDictRequest
	}

	// IListDictRequestBuilder 获取字典列表请求参数构造器
	IListDictRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.QueryDictListParams
	}

	listDictRequestBuilder struct {
		ctx context.Context
		*dictapi.ListDictRequest
	}

	// IUpdateDictStatusParamsBuilder 更新字典状态请求参数构造器
	IUpdateDictStatusParamsBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateDictStatusParams
	}

	updateDictStatusParamsBuilder struct {
		ctx context.Context
		*dictapi.BatchUpdateDictStatusRequest
	}

	// IDoDictBuilder 字典条目构造器
	IDoDictBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(imodel.IDict) *adminapi.DictItem
		// ToAPIs 转换为API对象列表
		ToAPIs([]imodel.IDict) []*adminapi.DictItem
		// ToSelect 转换为选择对象
		ToSelect(imodel.IDict) *adminapi.SelectItem
		// ToSelects 转换为选择对象列表
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
	userMap := getUsers(d.ctx, dict.GetCreatorID())
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
		Creator:      userMap[dict.GetCreatorID()],
	}
}

func (d *doDictBuilder) ToAPIs(dictList []imodel.IDict) []*adminapi.DictItem {
	if types.IsNil(d) || types.IsNil(dictList) {
		return nil
	}

	return types.SliceTo(dictList, func(item imodel.IDict) *adminapi.DictItem {
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

func (d *doDictBuilder) ToSelects(dictList []imodel.IDict) []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(dictList) {
		return nil
	}

	return types.SliceTo(dictList, func(item imodel.IDict) *adminapi.SelectItem {
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
		UpdateParam: NewParamsBuild(u.ctx).DictModuleBuilder().WithCreateDictRequest(u.GetData()).ToBo(),
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
