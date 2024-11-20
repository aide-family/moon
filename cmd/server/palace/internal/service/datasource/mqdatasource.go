package datasource

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	mqapi "github.com/aide-family/moon/api/admin/datasource"
)

// MqDatasourceService mq datasource service
type MqDatasourceService struct {
	mqapi.UnimplementedMqDatasourceServer

	mqDataSourceBiz *biz.MqDataSourceBiz
}

// NewMqDatasourceService new mq datasource service
func NewMqDatasourceService(mqDataSourceBiz *biz.MqDataSourceBiz) *MqDatasourceService {
	return &MqDatasourceService{mqDataSourceBiz: mqDataSourceBiz}
}

func (s *MqDatasourceService) CreateMqDatasource(ctx context.Context, req *mqapi.CreateMqDatasourceRequest) (*mqapi.CreateMqDatasourceReply, error) {
	params := builder.NewParamsBuild(ctx).MqDataSourceModuleBuild().WithCreateMqDatasourceRequest(req).ToBo()
	err := s.mqDataSourceBiz.CreateMqDataSource(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &mqapi.CreateMqDatasourceReply{}, nil
}
func (s *MqDatasourceService) UpdateMqDatasource(ctx context.Context, req *mqapi.UpdateMqDatasourceRequest) (*mqapi.UpdateMqDatasourceReply, error) {
	params := builder.NewParamsBuild(ctx).MqDataSourceModuleBuild().WithUpdateMqDatasourceRequest(req).ToBo()
	err := s.mqDataSourceBiz.UpdateMqDataSource(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &mqapi.UpdateMqDatasourceReply{}, nil
}
func (s *MqDatasourceService) DeleteMqDatasource(ctx context.Context, req *mqapi.DeleteMqDatasourceRequest) (*mqapi.DeleteMqDatasourceReply, error) {
	err := s.mqDataSourceBiz.DeleteMqDatasource(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &mqapi.DeleteMqDatasourceReply{}, nil
}
func (s *MqDatasourceService) GetMqDatasource(ctx context.Context, req *mqapi.GetMqDatasourceRequest) (*mqapi.GetMqDatasourceReply, error) {
	dataSource, err := s.mqDataSourceBiz.GetMqDataSource(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &mqapi.GetMqDatasourceReply{
		Detail: builder.NewParamsBuild(ctx).MqDataSourceModuleBuild().DoDataSourceModuleBuild().ToAPI(dataSource),
	}, nil
}
func (s *MqDatasourceService) ListMqDatasource(ctx context.Context, req *mqapi.ListMqDatasourceRequest) (*mqapi.ListMqDatasourceReply, error) {
	params := builder.NewParamsBuild(ctx).MqDataSourceModuleBuild().WithIListMqDatasourceRequest(req).ToBo()
	sourceList, err := s.mqDataSourceBiz.MqDataSourceList(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &mqapi.ListMqDatasourceReply{
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(params.Page),
		List:       builder.NewParamsBuild(ctx).MqDataSourceModuleBuild().DoDataSourceModuleBuild().ToAPIs(sourceList),
	}, nil
}
func (s *MqDatasourceService) UpdateMqDatasourceStatus(ctx context.Context, req *mqapi.UpdateMqDatasourceStatusRequest) (*mqapi.UpdateMqDatasourceStatusReply, error) {
	if err := s.mqDataSourceBiz.UpdateMqDataSourceStatus(ctx, &bo.UpdateMqDatasourceStatusParams{ID: req.GetId(), Status: vobj.Status(req.GetStatus())}); !types.IsNil(err) {
		return nil, err
	}
	return &mqapi.UpdateMqDatasourceStatusReply{}, nil
}
func (s *MqDatasourceService) GetMqDatasourceSelect(ctx context.Context, req *mqapi.GetMqDatasourceSelectRequest) (*mqapi.GetMqDatasourceSelectReply, error) {
	params := builder.NewParamsBuild(ctx).MqDataSourceModuleBuild().WithDatasourceSelectRequest(req).ToBo()
	datasourceSelect, err := s.mqDataSourceBiz.GetMqDatasourceSelect(ctx, params)
	if err != nil {
		return nil, err
	}
	return &mqapi.GetMqDatasourceSelectReply{
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(params.Page),
		List:       builder.NewParamsBuild(ctx).MqDataSourceModuleBuild().DoDataSourceModuleBuild().ToSelects(datasourceSelect),
	}, nil
}
