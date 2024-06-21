package datasource

import (
	"context"

	"github.com/aide-family/moon/api/admin"
	datasourceapi "github.com/aide-family/moon/api/admin/datasource"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type MetricService struct {
	datasourceapi.UnimplementedMetricServer

	metricBiz *biz.MetricBiz
}

func NewMetricService(metricBiz *biz.MetricBiz) *MetricService {
	return &MetricService{
		metricBiz: metricBiz,
	}
}

func (s *MetricService) UpdateMetric(ctx context.Context, req *datasourceapi.UpdateMetricRequest) (*datasourceapi.UpdateMetricReply, error) {
	params := &bo.UpdateMetricParams{
		ID:     req.GetId(),
		Unit:   req.GetUnit(),
		Remark: req.GetRemark(),
	}
	if err := s.metricBiz.UpdateMetricByID(ctx, params); err != nil {
		return nil, err
	}
	return &datasourceapi.UpdateMetricReply{}, nil
}

func (s *MetricService) GetMetric(ctx context.Context, req *datasourceapi.GetMetricRequest) (*datasourceapi.GetMetricReply, error) {
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
	return &datasourceapi.GetMetricReply{
		Data:       build.NewDatasourceMetricBuilder(detail).ToApi(),
		LabelCount: labelCount,
	}, nil
}

func (s *MetricService) ListMetric(ctx context.Context, req *datasourceapi.ListMetricRequest) (*datasourceapi.ListMetricReply, error) {
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
	return &datasourceapi.ListMetricReply{
		Pagination: build.NewPageBuilder(params.Page).ToApi(),
		List: types.SliceTo(list, func(item *bizmodel.DatasourceMetric) *admin.MetricDetail {
			return build.NewDatasourceMetricBuilder(item).ToApi()
		}),
	}, nil
}

func (s *MetricService) SelectMetric(ctx context.Context, req *datasourceapi.SelectMetricRequest) (*datasourceapi.SelectMetricReply, error) {
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

	return &datasourceapi.SelectMetricReply{
		Pagination: build.NewPageBuilder(params.Page).ToApi(),
		List: types.SliceTo(list, func(item *bo.SelectOptionBo) *admin.Select {
			return build.NewSelectBuilder(item).ToApi()
		}),
	}, nil
}

func (s *MetricService) DeleteMetric(ctx context.Context, req *datasourceapi.DeleteMetricRequest) (*datasourceapi.DeleteMetricReply, error) {
	if err := s.metricBiz.DeleteMetricByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &datasourceapi.DeleteMetricReply{}, nil
}
