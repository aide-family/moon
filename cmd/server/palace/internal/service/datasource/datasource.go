package datasource

import (
	"context"
	"encoding/json"

	"github.com/aide-cloud/moon/api"
	"github.com/aide-cloud/moon/api/admin"
	pb "github.com/aide-cloud/moon/api/admin/datasource"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/log"
)

type Service struct {
	pb.UnimplementedDatasourceServer

	datasourceBiz *biz.DatasourceBiz
}

func NewDatasourceService(datasourceBiz *biz.DatasourceBiz) *Service {
	return &Service{
		datasourceBiz: datasourceBiz,
	}
}

func (s *Service) CreateDatasource(ctx context.Context, req *pb.CreateDatasourceRequest) (*pb.CreateDatasourceReply, error) {
	configBytes, _ := json.Marshal(req.GetConfig())
	params := &bo.CreateDatasourceParams{
		Name:        req.GetName(),
		Type:        vobj.DatasourceType(req.GetType()),
		Endpoint:    req.GetEndpoint(),
		Status:      vobj.Status(req.GetStatus()),
		Remark:      req.GetRemark(),
		Config:      string(configBytes),
		StorageType: vobj.StorageType(req.GetStorageType()),
	}
	datasourceDetail, err := s.datasourceBiz.CreateDatasource(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	// 记录操作日志
	log.Debugw("datasourceDetail", datasourceDetail)
	return &pb.CreateDatasourceReply{}, nil
}

func (s *Service) UpdateDatasource(ctx context.Context, req *pb.UpdateDatasourceRequest) (*pb.UpdateDatasourceReply, error) {
	data := req.GetData()
	params := &bo.UpdateDatasourceBaseInfoParams{
		ID:     req.GetId(),
		Name:   data.GetName(),
		Status: vobj.Status(data.GetStatus()),
		Remark: data.GetRemark(),
	}
	if err := s.datasourceBiz.UpdateDatasourceBaseInfo(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &pb.UpdateDatasourceReply{}, nil
}

func (s *Service) DeleteDatasource(ctx context.Context, req *pb.DeleteDatasourceRequest) (*pb.DeleteDatasourceReply, error) {
	if err := s.datasourceBiz.DeleteDatasource(ctx, req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &pb.DeleteDatasourceReply{}, nil
}

func (s *Service) GetDatasource(ctx context.Context, req *pb.GetDatasourceRequest) (*pb.GetDatasourceReply, error) {
	datasourceDetail, err := s.datasourceBiz.GetDatasource(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.GetDatasourceReply{
		Data: build.NewDatasourceBuild(datasourceDetail).ToApi(),
	}, nil
}

func (s *Service) ListDatasource(ctx context.Context, req *pb.ListDatasourceRequest) (*pb.ListDatasourceReply, error) {
	params := &bo.QueryDatasourceListParams{
		Page:        types.NewPagination(req.GetPagination()),
		Keyword:     req.GetKeyword(),
		Type:        vobj.DatasourceType(req.GetType()),
		Status:      vobj.Status(req.GetStatus()),
		StorageType: vobj.StorageType(req.GetStorageType()),
	}
	datasourceList, err := s.datasourceBiz.ListDatasource(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.ListDatasourceReply{
		Pagination: build.NewPageBuild(params.Page).ToApi(),
		Data: types.SliceTo(datasourceList, func(item *bizmodel.Datasource) *admin.Datasource {
			return build.NewDatasourceBuild(item).ToApi()
		}),
	}, nil
}

func (s *Service) UpdateDatasourceStatus(ctx context.Context, req *pb.UpdateDatasourceStatusRequest) (*pb.UpdateDatasourceStatusReply, error) {
	if err := s.datasourceBiz.UpdateDatasourceStatus(ctx, vobj.Status(req.GetStatus()), req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &pb.UpdateDatasourceStatusReply{}, nil
}

func (s *Service) GetDatasourceSelect(ctx context.Context, req *pb.GetDatasourceSelectRequest) (*pb.GetDatasourceSelectReply, error) {
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
	return &pb.GetDatasourceSelectReply{
		Data: types.SliceTo(list, func(item *bo.SelectOptionBo) *admin.Select {
			return build.NewSelectBuild(item).ToApi()
		}),
	}, nil
}

// SyncDatasourceMeta 同步数据源元数据
func (s *Service) SyncDatasourceMeta(ctx context.Context, req *pb.SyncDatasourceMetaRequest) (*pb.SyncDatasourceMetaReply, error) {
	if err := s.datasourceBiz.SyncDatasourceMeta(ctx, req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &pb.SyncDatasourceMetaReply{}, nil
}

// DatasourceQuery 查询数据
func (s *Service) DatasourceQuery(ctx context.Context, req *pb.DatasourceQueryRequest) (*pb.DatasourceQueryReply, error) {
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
	return &pb.DatasourceQueryReply{
		List: types.SliceTo(query, func(item *bo.DatasourceQueryData) *api.MetricQueryResult {
			return build.NewDatasourceQueryDataBuild(item).ToApi()
		}),
	}, nil
}
