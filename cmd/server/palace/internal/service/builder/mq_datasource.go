package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	mqapi "github.com/aide-family/moon/api/admin/datasource"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ IMqDataSourceModuleBuild = (*mqDatasourceModuleBuild)(nil)

type (
	mqDatasourceModuleBuild struct {
		ctx context.Context
	}

	IMqDataSourceModuleBuild interface {
		DoDataSourceModuleBuild() IDoMqDataSourceBuild
		WithCreateMqDatasourceRequest(*mqapi.CreateMqDatasourceRequest) ICreateMqCreateDatasourceRequestBuilder
		WithIListMqDatasourceRequest(*mqapi.ListMqDatasourceRequest) IListMqDatasourceRequest
		WithUpdateMqDatasourceRequest(*mqapi.UpdateMqDatasourceRequest) IUpdateMqDatasourceRequestBuilder
		WithDatasourceSelectRequest(request *mqapi.GetMqDatasourceSelectRequest) IDatasourceSelectBuilder
	}

	IDoMqDataSourceBuild interface {
		ToAPI(*bizmodel.MqDatasource, ...map[uint32]*adminapi.UserItem) *adminapi.MqDatasourceItem
		ToAPIs([]*bizmodel.MqDatasource) []*adminapi.MqDatasourceItem
		ToSelect(*bizmodel.MqDatasource) *adminapi.SelectItem
		ToSelects([]*bizmodel.MqDatasource) []*adminapi.SelectItem
	}

	doMqDataSourceBuild struct {
		ctx context.Context
	}

	ICreateMqCreateDatasourceRequestBuilder interface {
		ToBo() *bo.CreateMqDatasourceParams
	}

	createMqCreateDatasourceRequestBuilder struct {
		ctx     context.Context
		request *mqapi.CreateMqDatasourceRequest
	}
	IUpdateMqDatasourceRequestBuilder interface {
		ToBo() *bo.UpdateMqDatasourceParams
	}

	updateMqDatasourceRequestBuilder struct {
		ctx     context.Context
		request *mqapi.UpdateMqDatasourceRequest
	}

	IListMqDatasourceRequest interface {
		ToBo() *bo.QueryMqDatasourceListParams
	}
	listMqDatasourceRequest struct {
		ctx     context.Context
		request *mqapi.ListMqDatasourceRequest
	}

	IDatasourceSelectBuilder interface {
		ToBo() *bo.QueryMqDatasourceListParams
	}
	datasourceSelectBuilder struct {
		ctx     context.Context
		request *mqapi.GetMqDatasourceSelectRequest
	}
)

func (m *datasourceSelectBuilder) ToBo() *bo.QueryMqDatasourceListParams {
	if types.IsNil(m) || types.IsNil(m.request) {
		return nil
	}
	request := m.request
	return &bo.QueryMqDatasourceListParams{
		Page:           types.NewPagination(request.GetPagination()),
		Keyword:        request.GetKeyword(),
		DatasourceType: vobj.DatasourceType(request.GetDatasourceType()),
		StorageType:    vobj.StorageType(request.GetStorageType()),
		Status:         vobj.Status(request.GetStatus()),
	}
}

func (m *mqDatasourceModuleBuild) WithDatasourceSelectRequest(request *mqapi.GetMqDatasourceSelectRequest) IDatasourceSelectBuilder {
	return &datasourceSelectBuilder{
		ctx:     m.ctx,
		request: request,
	}
}

func (d *doMqDataSourceBuild) ToAPI(datasource *bizmodel.MqDatasource, userMaps ...map[uint32]*adminapi.UserItem) *adminapi.MqDatasourceItem {
	if types.IsNil(datasource) || types.IsNil(d) {
		return nil
	}
	userMap := getUsers(d.ctx, userMaps, datasource.CreatorID)
	config := make(map[string]string)
	_ = types.Unmarshal([]byte(datasource.Config), &config)
	return &adminapi.MqDatasourceItem{
		Id:             datasource.ID,
		Name:           datasource.Name,
		DatasourceType: api.DatasourceType(datasource.DatasourceType),
		Endpoint:       datasource.Endpoint,
		Status:         api.Status(datasource.Status),
		Remark:         datasource.Remark,
		CreatedAt:      datasource.CreatedAt.String(),
		UpdatedAt:      datasource.UpdatedAt.String(),
		Config:         config,
		Creator:        userMap[datasource.CreatorID],
		StorageType:    api.StorageType(datasource.StorageType),
	}
}

func (d *doMqDataSourceBuild) ToAPIs(datasources []*bizmodel.MqDatasource) []*adminapi.MqDatasourceItem {
	if types.IsNil(datasources) || types.IsNil(d) {
		return nil
	}
	ids := types.SliceTo(datasources, func(item *bizmodel.MqDatasource) uint32 {
		return item.CreatorID
	})
	userMap := getUsers(d.ctx, nil, ids...)
	return types.SliceTo(datasources, func(item *bizmodel.MqDatasource) *adminapi.MqDatasourceItem {
		return d.ToAPI(item, userMap)
	})
}

func (d *doMqDataSourceBuild) ToSelect(datasource *bizmodel.MqDatasource) *adminapi.SelectItem {
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

func (d *doMqDataSourceBuild) ToSelects(datasources []*bizmodel.MqDatasource) []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(datasources) {
		return nil
	}
	return types.SliceTo(datasources, func(item *bizmodel.MqDatasource) *adminapi.SelectItem {
		return d.ToSelect(item)
	})
}

func (m *mqDatasourceModuleBuild) DoDataSourceModuleBuild() IDoMqDataSourceBuild {
	return &doMqDataSourceBuild{ctx: m.ctx}
}

func (m *updateMqDatasourceRequestBuilder) ToBo() *bo.UpdateMqDatasourceParams {
	if types.IsNil(m) || types.IsNil(m.request) {
		return nil
	}
	request := m.request
	updateParam := m.request.GetUpdateParam()
	return &bo.UpdateMqDatasourceParams{
		ID: request.GetId(),
		UpdateParam: &bo.CreateMqDatasourceParams{
			Name:           updateParam.GetName(),
			DatasourceType: vobj.DatasourceType(updateParam.GetDatasourceType()),
			StorageType:    vobj.StorageType(updateParam.GetStorageType()),
			Endpoint:       updateParam.GetEndpoint(),
			Config:         updateParam.GetConfig(),
			Remark:         updateParam.GetRemark(),
			Status:         vobj.Status(updateParam.GetStatus()),
		},
	}
}

func (m *listMqDatasourceRequest) ToBo() *bo.QueryMqDatasourceListParams {
	if types.IsNil(m) || types.IsNil(m.request) {
		return nil
	}
	request := m.request
	return &bo.QueryMqDatasourceListParams{
		Page:           types.NewPagination(request.GetPagination()),
		Keyword:        request.GetKeyword(),
		DatasourceType: vobj.DatasourceType(request.GetDatasourceType()),
		StorageType:    vobj.StorageType(request.GetStorageType()),
		Status:         vobj.Status(request.GetStatus()),
	}
}

func (m *createMqCreateDatasourceRequestBuilder) ToBo() *bo.CreateMqDatasourceParams {
	if types.IsNil(m) || types.IsNil(m.request) {
		return nil
	}
	request := m.request
	return &bo.CreateMqDatasourceParams{
		Name:           request.GetName(),
		DatasourceType: vobj.DatasourceType(request.GetDatasourceType()),
		StorageType:    vobj.StorageType(request.GetStorageType()),
		Endpoint:       request.GetEndpoint(),
		Config:         request.GetConfig(),
		Remark:         request.GetRemark(),
		Status:         vobj.Status(request.GetStatus()),
	}
}

func (m *mqDatasourceModuleBuild) WithCreateMqDatasourceRequest(request *mqapi.CreateMqDatasourceRequest) ICreateMqCreateDatasourceRequestBuilder {
	return &createMqCreateDatasourceRequestBuilder{
		request: request,
		ctx:     m.ctx,
	}
}

func (m *mqDatasourceModuleBuild) WithIListMqDatasourceRequest(request *mqapi.ListMqDatasourceRequest) IListMqDatasourceRequest {
	return &listMqDatasourceRequest{
		request: request,
		ctx:     m.ctx,
	}
}

func (m *mqDatasourceModuleBuild) WithUpdateMqDatasourceRequest(request *mqapi.UpdateMqDatasourceRequest) IUpdateMqDatasourceRequestBuilder {
	return &updateMqDatasourceRequestBuilder{
		request: request,
		ctx:     m.ctx,
	}
}
