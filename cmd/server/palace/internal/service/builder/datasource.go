package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	datasourceapi "github.com/aide-family/moon/api/admin/datasource"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ IDatasourceModuleBuilder = (*datasourceModuleBuilder)(nil)

type (
	datasourceModuleBuilder struct {
		ctx context.Context
	}

	// IDatasourceModuleBuilder 数据源模块构造器
	IDatasourceModuleBuilder interface {
		// DoDatasourceBuilder 数据源条目构造器
		DoDatasourceBuilder() IDoDatasourceBuilder

		// WithCreateDatasourceRequest 创建数据源请求参数构造器
		WithCreateDatasourceRequest(*datasourceapi.CreateDatasourceRequest) ICreateDatasourceRequestBuilder

		// WithUpdateDatasourceRequest 更新数据源请求参数构造器
		WithUpdateDatasourceRequest(*datasourceapi.UpdateDatasourceRequest) IUpdateDatasourceRequestBuilder

		// WithListDatasourceRequest 获取数据源列表请求参数构造器
		WithListDatasourceRequest(*datasourceapi.ListDatasourceRequest) IListDatasourceRequestBuilder

		// DatasourceBuilder 业务对象构造器
		DatasourceBuilder() IDatasourceBuilder
	}

	// IDoDatasourceBuilder 数据源条目构造器
	IDoDatasourceBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*bizmodel.Datasource) *adminapi.DatasourceItem

		// ToAPIs 转换为API对象列表
		ToAPIs([]*bizmodel.Datasource) []*adminapi.DatasourceItem

		// ToSelect 转换为选择对象
		ToSelect(*bizmodel.Datasource) *adminapi.SelectItem

		// ToSelects 转换为选择对象列表
		ToSelects([]*bizmodel.Datasource) []*adminapi.SelectItem
	}

	doDatasourceBuilder struct {
		ctx context.Context
	}

	// ICreateDatasourceRequestBuilder 创建数据源请求参数构造器
	ICreateDatasourceRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.CreateDatasourceParams
	}

	createDatasourceRequestBuilder struct {
		ctx context.Context
		*datasourceapi.CreateDatasourceRequest
	}

	// IUpdateDatasourceRequestBuilder 更新数据源请求参数构造器
	IUpdateDatasourceRequestBuilder interface {
		ToBo() *bo.UpdateDatasourceBaseInfoParams
	}

	updateDatasourceRequestBuilder struct {
		ctx context.Context
		*datasourceapi.UpdateDatasourceRequest
	}

	// IListDatasourceRequestBuilder 获取数据源列表请求参数构造器
	IListDatasourceRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.QueryDatasourceListParams
	}

	listDatasourceRequestBuilder struct {
		ctx context.Context
		*datasourceapi.ListDatasourceRequest
	}

	// IDatasourceBuilder 业务对象构造器
	IDatasourceBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*bizmodel.Datasource) *api.DatasourceItem
		// ToAPIs 转换为API对象列表
		ToAPIs([]*bizmodel.Datasource) []*api.DatasourceItem
	}

	boDatasourceBuilder struct {
		ctx context.Context
	}
)

func (b *boDatasourceBuilder) ToAPI(datasource *bizmodel.Datasource) *api.DatasourceItem {
	if types.IsNil(b) || types.IsNil(datasource) {
		return nil
	}
	return &api.DatasourceItem{
		Category:    api.DatasourceType(datasource.Category),
		StorageType: api.StorageType(datasource.StorageType),
		Config:      datasource.Config.String(),
		Endpoint:    datasource.Endpoint,
		Id:          datasource.ID,
		Status:      api.Status(datasource.Status),
		TeamId:      datasource.TeamID,
	}
}

func (b *boDatasourceBuilder) ToAPIs(datasource []*bizmodel.Datasource) []*api.DatasourceItem {
	return types.SliceTo(datasource, func(item *bizmodel.Datasource) *api.DatasourceItem {
		return b.ToAPI(item)
	})
}

func (d *datasourceModuleBuilder) DatasourceBuilder() IDatasourceBuilder {
	return &boDatasourceBuilder{ctx: d.ctx}
}

func (l *listDatasourceRequestBuilder) ToBo() *bo.QueryDatasourceListParams {
	if types.IsNil(l) || types.IsNil(l.ListDatasourceRequest) {
		return nil
	}

	return &bo.QueryDatasourceListParams{
		Page:           types.NewPagination(l.GetPagination()),
		Keyword:        l.GetKeyword(),
		DatasourceType: vobj.DatasourceType(l.GetDatasourceType()),
		Status:         vobj.Status(l.GetStatus()),
		StorageType:    vobj.StorageType(l.GetStorageType()),
	}
}

func (u *updateDatasourceRequestBuilder) ToBo() *bo.UpdateDatasourceBaseInfoParams {
	if types.IsNil(u) || types.IsNil(u.UpdateDatasourceRequest) {
		return nil
	}

	return &bo.UpdateDatasourceBaseInfoParams{
		ID:             u.GetId(),
		Name:           u.GetName(),
		Endpoint:       u.GetEndpoint(),
		Status:         vobj.Status(u.GetStatus()),
		Config:         datasource.NewDatasourceConfigByString(u.GetConfig()),
		Remark:         u.GetRemark(),
		StorageType:    vobj.StorageType(u.GetStorageType()),
		DatasourceType: vobj.DatasourceType(u.GetDatasourceType()),
	}
}

func (c *createDatasourceRequestBuilder) ToBo() *bo.CreateDatasourceParams {
	if types.IsNil(c) || types.IsNil(c.CreateDatasourceRequest) {
		return nil
	}

	return &bo.CreateDatasourceParams{
		Name:           c.GetName(),
		DatasourceType: vobj.DatasourceType(c.GetDatasourceType()),
		Endpoint:       c.GetEndpoint(),
		Status:         vobj.Status(c.GetStatus()),
		Remark:         c.GetRemark(),
		Config:         datasource.NewDatasourceConfigByString(c.GetConfig()),
		StorageType:    vobj.StorageType(c.GetStorageType()),
	}
}

func (d *doDatasourceBuilder) ToAPI(datasource *bizmodel.Datasource) *adminapi.DatasourceItem {
	if types.IsNil(datasource) || types.IsNil(d) {
		return nil
	}

	userMap := getUsers(d.ctx, datasource.CreatorID)
	return &adminapi.DatasourceItem{
		Id:             datasource.ID,
		Name:           datasource.Name,
		DatasourceType: api.DatasourceType(datasource.Category),
		Endpoint:       datasource.Endpoint,
		Status:         api.Status(datasource.Status),
		CreatedAt:      datasource.CreatedAt.String(),
		UpdatedAt:      datasource.UpdatedAt.String(),
		Config:         datasource.Config.String(),
		Remark:         datasource.Remark,
		StorageType:    api.StorageType(datasource.StorageType),
		Creator:        userMap[datasource.CreatorID],
	}
}

func (d *doDatasourceBuilder) ToAPIs(datasources []*bizmodel.Datasource) []*adminapi.DatasourceItem {
	if types.IsNil(datasources) || types.IsNil(d) {
		return nil
	}

	return types.SliceTo(datasources, func(item *bizmodel.Datasource) *adminapi.DatasourceItem {
		return d.ToAPI(item)
	})
}

func (d *doDatasourceBuilder) ToSelect(datasource *bizmodel.Datasource) *adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(datasource) {
		return nil
	}

	return &adminapi.SelectItem{
		Value:    datasource.ID,
		Label:    datasource.Name,
		Children: nil,
		Disabled: datasource.DeletedAt > 0 || !datasource.Status.IsEnable(),
		Extend: &adminapi.SelectExtend{
			Remark: datasource.Remark,
		},
	}
}

func (d *doDatasourceBuilder) ToSelects(datasources []*bizmodel.Datasource) []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(datasources) {
		return nil
	}

	return types.SliceTo(datasources, func(item *bizmodel.Datasource) *adminapi.SelectItem {
		return d.ToSelect(item)
	})
}

func (d *datasourceModuleBuilder) DoDatasourceBuilder() IDoDatasourceBuilder {
	return &doDatasourceBuilder{ctx: d.ctx}
}

func (d *datasourceModuleBuilder) WithCreateDatasourceRequest(request *datasourceapi.CreateDatasourceRequest) ICreateDatasourceRequestBuilder {
	return &createDatasourceRequestBuilder{ctx: d.ctx, CreateDatasourceRequest: request}
}

func (d *datasourceModuleBuilder) WithUpdateDatasourceRequest(request *datasourceapi.UpdateDatasourceRequest) IUpdateDatasourceRequestBuilder {
	return &updateDatasourceRequestBuilder{ctx: d.ctx, UpdateDatasourceRequest: request}
}

func (d *datasourceModuleBuilder) WithListDatasourceRequest(request *datasourceapi.ListDatasourceRequest) IListDatasourceRequestBuilder {
	return &listDatasourceRequestBuilder{ctx: d.ctx, ListDatasourceRequest: request}
}
