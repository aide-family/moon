package datasource

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	datasourceapi "github.com/aide-family/moon/api/admin/datasource"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/log"
)

type Service struct {
	datasourceapi.UnimplementedDatasourceServer

	datasourceBiz *biz.DatasourceBiz
}

func NewDatasourceService(datasourceBiz *biz.DatasourceBiz) *Service {
	return &Service{
		datasourceBiz: datasourceBiz,
	}
}

func (s *Service) CreateDatasource(ctx context.Context, req *datasourceapi.CreateDatasourceRequest) (*datasourceapi.CreateDatasourceReply, error) {
	configBytes, _ := json.Marshal(req.GetConfig())
	params := build.NewBuilder().WithCreateDatasourceBo(req).ToCreateDatasourceBO(configBytes)
	datasourceDetail, err := s.datasourceBiz.CreateDatasource(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	// 记录操作日志
	log.Debugw("datasourceDetail", datasourceDetail)
	return &datasourceapi.CreateDatasourceReply{}, nil
}

func (s *Service) UpdateDatasource(ctx context.Context, req *datasourceapi.UpdateDatasourceRequest) (*datasourceapi.UpdateDatasourceReply, error) {
	params := build.NewBuilder().WithUpdateDatasourceBo(req).ToUpdateDatasourceBO()
	if err := s.datasourceBiz.UpdateDatasourceBaseInfo(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.UpdateDatasourceReply{}, nil
}

func (s *Service) DeleteDatasource(ctx context.Context, req *datasourceapi.DeleteDatasourceRequest) (*datasourceapi.DeleteDatasourceReply, error) {
	if err := s.datasourceBiz.DeleteDatasource(ctx, req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.DeleteDatasourceReply{}, nil
}

func (s *Service) GetDatasource(ctx context.Context, req *datasourceapi.GetDatasourceRequest) (*datasourceapi.GetDatasourceReply, error) {
	datasourceDetail, err := s.datasourceBiz.GetDatasource(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.GetDatasourceReply{
		Data: build.NewBuilder().WithContext(ctx).WithDoDatasource(datasourceDetail).ToApi(),
	}, nil
}

func (s *Service) ListDatasource(ctx context.Context, req *datasourceapi.ListDatasourceRequest) (*datasourceapi.ListDatasourceReply, error) {
	params := build.NewBuilder().WithListDatasourceBo(req).ToListDatasourceBo()
	datasourceList, err := s.datasourceBiz.ListDatasource(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.ListDatasourceReply{
		Pagination: build.NewPageBuilder(params.Page).ToApi(),
		List: types.SliceTo(datasourceList, func(item *bizmodel.Datasource) *admin.Datasource {
			return build.NewBuilder().WithContext(ctx).WithDoDatasource(item).ToApi()
		}),
	}, nil
}

func (s *Service) UpdateDatasourceStatus(ctx context.Context, req *datasourceapi.UpdateDatasourceStatusRequest) (*datasourceapi.UpdateDatasourceStatusReply, error) {
	if err := s.datasourceBiz.UpdateDatasourceStatus(ctx, vobj.Status(req.GetStatus()), req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.UpdateDatasourceStatusReply{}, nil
}

func (s *Service) GetDatasourceSelect(ctx context.Context, req *datasourceapi.GetDatasourceSelectRequest) (*datasourceapi.GetDatasourceSelectReply, error) {
	params := &bo.QueryDatasourceListParams{
		Page:        nil,
		Keyword:     req.GetKeyword(),
		Type:        vobj.DatasourceType(req.GetType()),
		Status:      vobj.Status(req.GetStatus()),
		StorageType: vobj.StorageType(req.GetStorageType()),
	}
	list, err := s.datasourceBiz.GetDatasourceSelect(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.GetDatasourceSelectReply{
		Data: types.SliceTo(list, func(item *bo.SelectOptionBo) *admin.Select {
			return build.NewSelectBuilder(item).ToApi()
		}),
	}, nil
}

// SyncDatasourceMeta 同步数据源元数据
func (s *Service) SyncDatasourceMeta(ctx context.Context, req *datasourceapi.SyncDatasourceMetaRequest) (*datasourceapi.SyncDatasourceMetaReply, error) {
	if err := s.datasourceBiz.SyncDatasourceMeta(ctx, req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.SyncDatasourceMetaReply{}, nil
}

// DatasourceQuery 查询数据
func (s *Service) DatasourceQuery(ctx context.Context, req *datasourceapi.DatasourceQueryRequest) (*datasourceapi.DatasourceQueryReply, error) {
	params := &bo.DatasourceQueryParams{
		DatasourceID: req.GetId(),
		Query:        req.GetQuery(),
		Step:         req.GetStep(),
		TimeRange:    req.GetRange(),
	}
	query, err := s.datasourceBiz.Query(ctx, params)
	if err != nil {
		return nil, err
	}
	return &datasourceapi.DatasourceQueryReply{
		List: types.SliceTo(query, func(item *bo.DatasourceQueryData) *api.MetricQueryResult {
			return build.NewBuilder().WithBoDatasourceQueryData(item).ToApi()
		}),
	}, nil
}
