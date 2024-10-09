package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	datasourceapi "github.com/aide-family/moon/api/admin/datasource"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ IDatasourceModuleBuilder = (*datasourceModuleBuilder)(nil)

type (
	datasourceModuleBuilder struct {
		ctx context.Context
	}

	IDatasourceModuleBuilder interface {
		DoDatasourceBuilder() IDoDatasourceBuilder
		WithCreateDatasourceRequest(*datasourceapi.CreateDatasourceRequest) ICreateDatasourceRequestBuilder
		WithUpdateDatasourceRequest(*datasourceapi.UpdateDatasourceRequest) IUpdateDatasourceRequestBuilder
		WithListDatasourceRequest(*datasourceapi.ListDatasourceRequest) IListDatasourceRequestBuilder

		BoDatasourceBuilder() IBoDatasourceBuilder
	}

	IDoDatasourceBuilder interface {
		ToAPI(*bizmodel.Datasource) *adminapi.DatasourceItem
		ToAPIs([]*bizmodel.Datasource) []*adminapi.DatasourceItem
		ToBo(*bizmodel.Datasource) *bo.Datasource
		ToBos([]*bizmodel.Datasource) []*bo.Datasource
		ToSelect(*bizmodel.Datasource) *adminapi.SelectItem
		ToSelects([]*bizmodel.Datasource) []*adminapi.SelectItem
	}

	doDatasourceBuilder struct {
		ctx context.Context
	}

	ICreateDatasourceRequestBuilder interface {
		ToBo() *bo.CreateDatasourceParams
	}

	createDatasourceRequestBuilder struct {
		ctx context.Context
		*datasourceapi.CreateDatasourceRequest
	}

	IUpdateDatasourceRequestBuilder interface {
		ToBo() *bo.UpdateDatasourceBaseInfoParams
	}

	updateDatasourceRequestBuilder struct {
		ctx context.Context
		*datasourceapi.UpdateDatasourceRequest
	}

	IListDatasourceRequestBuilder interface {
		ToBo() *bo.QueryDatasourceListParams
	}

	listDatasourceRequestBuilder struct {
		ctx context.Context
		*datasourceapi.ListDatasourceRequest
	}

	IBoDatasourceBuilder interface {
		ToAPI(*bo.Datasource) *api.Datasource
		ToAPIs([]*bo.Datasource) []*api.Datasource
	}

	boDatasourceBuilder struct {
		ctx context.Context
	}
)

func (d *doDatasourceBuilder) ToBo(datasource *bizmodel.Datasource) *bo.Datasource {
	if types.IsNil(datasource) || types.IsNil(d) {
		return nil
	}

	config := make(map[string]string)
	_ = types.Unmarshal([]byte(datasource.Config), &config)
	return &bo.Datasource{
		Category:    datasource.Category,
		StorageType: datasource.StorageType,
		Config:      config,
		Endpoint:    datasource.Endpoint,
		ID:          datasource.ID,
	}
}

func (d *doDatasourceBuilder) ToBos(datasources []*bizmodel.Datasource) []*bo.Datasource {
	if types.IsNil(datasources) || types.IsNil(d) {
		return nil
	}

	return types.SliceTo(datasources, func(item *bizmodel.Datasource) *bo.Datasource {
		return d.ToBo(item)
	})
}

func (b *boDatasourceBuilder) ToAPI(datasource *bo.Datasource) *api.Datasource {
	if types.IsNil(datasource) || types.IsNil(b) {
		return nil
	}

	return &api.Datasource{
		Category:    api.DatasourceType(datasource.Category),
		StorageType: api.StorageType(datasource.StorageType),
		Config:      datasource.Config,
		Endpoint:    datasource.Endpoint,
		Id:          datasource.ID,
	}
}

func (b *boDatasourceBuilder) ToAPIs(datasources []*bo.Datasource) []*api.Datasource {
	if types.IsNil(datasources) || types.IsNil(b) {
		return nil
	}

	return types.SliceTo(datasources, func(item *bo.Datasource) *api.Datasource {
		return b.ToAPI(item)
	})
}

func (d *datasourceModuleBuilder) BoDatasourceBuilder() IBoDatasourceBuilder {
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
		Status:         vobj.Status(u.GetStatus()),
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
		Config:         c.GetConfig(),
		StorageType:    vobj.StorageType(c.GetStorageType()),
	}
}

func (d *doDatasourceBuilder) ToAPI(datasource *bizmodel.Datasource) *adminapi.DatasourceItem {
	if types.IsNil(datasource) || types.IsNil(d) {
		return nil
	}

	config := make(map[string]string)
	_ = types.Unmarshal([]byte(datasource.Config), &config)
	return &adminapi.DatasourceItem{
		Id:             datasource.ID,
		Name:           datasource.Name,
		DatasourceType: api.DatasourceType(datasource.Category),
		Endpoint:       datasource.Endpoint,
		Status:         api.Status(datasource.Status),
		CreatedAt:      datasource.CreatedAt.String(),
		UpdatedAt:      datasource.UpdatedAt.String(),
		Config:         config,
		Remark:         datasource.Remark,
		StorageType:    api.StorageType(datasource.StorageType),
		Creator:        nil, // TODO impl
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
