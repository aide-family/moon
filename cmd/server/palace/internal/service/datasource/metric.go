package datasource

import (
	"context"

	"github.com/aide-family/moon/api/admin"
	pb "github.com/aide-family/moon/api/admin/datasource"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/helper/model/bizmodel"
	"github.com/aide-family/moon/pkg/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type MetricService struct {
	pb.UnimplementedMetricServer

	metricBiz *biz.MetricBiz
}

func NewMetricService(metricBiz *biz.MetricBiz) *MetricService {
	return &MetricService{
		metricBiz: metricBiz,
	}
}

func (s *MetricService) UpdateMetric(ctx context.Context, req *pb.UpdateMetricRequest) (*pb.UpdateMetricReply, error) {
	params := &bo.UpdateMetricParams{
		ID:     req.GetId(),
		Unit:   req.GetUnit(),
		Remark: req.GetRemark(),
	}
	if err := s.metricBiz.UpdateMetricByID(ctx, params); err != nil {
		return nil, err
	}
	return &pb.UpdateMetricReply{}, nil
}

func (s *MetricService) GetMetric(ctx context.Context, req *pb.GetMetricRequest) (*pb.GetMetricReply, error) {
	params := &bo.GetMetricParams{
		ID:           req.GetId(),
		WithRelation: req.GetWithRelation(),
	}
	detail, err := s.metricBiz.GetMetricByID(ctx, params)
	if err != nil {
		return nil, err
	}
	labelCount, err := s.metricBiz.GetMetricLabelCount(ctx, params.ID)
	if err != nil {
		return nil, err
	}
	return &pb.GetMetricReply{
		Data:       build.NewDatasourceMetricBuild(detail).ToApi(),
		LabelCount: labelCount,
	}, nil
}

func (s *MetricService) ListMetric(ctx context.Context, req *pb.ListMetricRequest) (*pb.ListMetricReply, error) {
	params := &bo.QueryMetricListParams{
		Page:         types.NewPage(int(req.GetPagination().GetPageNum()), int(req.GetPagination().GetPageSize())),
		Keyword:      req.GetKeyword(),
		DatasourceID: req.GetDatasourceId(),
		MetricType:   vobj.MetricType(req.GetMetricType()),
	}
	list, err := s.metricBiz.ListMetric(ctx, params)
	if err != nil {
		return nil, err
	}
	return &pb.ListMetricReply{
		Pagination: build.NewPageBuild(params.Page).ToApi(),
		List: types.SliceTo(list, func(item *bizmodel.DatasourceMetric) *admin.MetricDetail {
			return build.NewDatasourceMetricBuild(item).ToApi()
		}),
	}, nil
}

func (s *MetricService) SelectMetric(ctx context.Context, req *pb.SelectMetricRequest) (*pb.SelectMetricReply, error) {
	params := &bo.QueryMetricListParams{
		Page:         types.NewPage(int(req.GetPagination().GetPageNum()), int(req.GetPagination().GetPageSize())),
		Keyword:      req.GetKeyword(),
		DatasourceID: req.GetDatasourceId(),
		MetricType:   vobj.MetricType(req.GetMetricType()),
	}
	list, err := s.metricBiz.SelectMetric(ctx, params)
	if err != nil {
		return nil, err
	}

	return &pb.SelectMetricReply{
		Pagination: build.NewPageBuild(params.Page).ToApi(),
		List: types.SliceTo(list, func(item *bo.SelectOptionBo) *admin.Select {
			return build.NewSelectBuild(item).ToApi()
		}),
	}, nil
}

func (s *MetricService) DeleteMetric(ctx context.Context, req *pb.DeleteMetricRequest) (*pb.DeleteMetricReply, error) {
	if err := s.metricBiz.DeleteMetricByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &pb.DeleteMetricReply{}, nil
}
