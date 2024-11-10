package datasource

import (
	"context"

	"github.com/aide-family/moon/api/admin"
	datasourceapi "github.com/aide-family/moon/api/admin/datasource"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// MetricService 指标服务
type MetricService struct {
	datasourceapi.UnimplementedMetricServer

	metricBiz *biz.MetricBiz
}

// NewMetricService 创建指标服务
func NewMetricService(metricBiz *biz.MetricBiz) *MetricService {
	return &MetricService{
		metricBiz: metricBiz,
	}
}

// UpdateMetric 更新指标
func (s *MetricService) UpdateMetric(ctx context.Context, req *datasourceapi.UpdateMetricRequest) (*datasourceapi.UpdateMetricReply, error) {
	params := builder.NewParamsBuild(ctx).MetricModuleBuilder().WithUpdateMetricRequest(req).ToBo()
	if err := s.metricBiz.UpdateMetricByID(ctx, params); err != nil {
		return nil, err
	}
	return &datasourceapi.UpdateMetricReply{}, nil
}

// GetMetric 获取指标
func (s *MetricService) GetMetric(ctx context.Context, req *datasourceapi.GetMetricRequest) (*datasourceapi.GetMetricReply, error) {
	params := builder.NewParamsBuild(ctx).MetricModuleBuilder().WithGetMetricRequest(req).ToBo()
	detail, err := s.metricBiz.GetMetricByID(ctx, params)
	if err != nil {
		return nil, err
	}
	labelCount, err := s.metricBiz.GetMetricLabelCount(ctx, params.ID)
	if err != nil {
		return nil, err
	}
	return &datasourceapi.GetMetricReply{
		Data:       builder.NewParamsBuild(ctx).MetricModuleBuilder().DoMetricBuilder().ToAPI(detail),
		LabelCount: labelCount,
	}, nil
}

// ListMetric 获取指标列表
func (s *MetricService) ListMetric(ctx context.Context, req *datasourceapi.ListMetricRequest) (*datasourceapi.ListMetricReply, error) {
	params := builder.NewParamsBuild(ctx).MetricModuleBuilder().WithListMetricRequest(req).ToBo()
	list, err := s.metricBiz.ListMetric(ctx, params)
	if err != nil {
		return nil, err
	}
	return &datasourceapi.ListMetricReply{
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(params.Page),
		List:       builder.NewParamsBuild(ctx).MetricModuleBuilder().DoMetricBuilder().ToAPIs(list),
	}, nil
}

// SelectMetric 获取指标下拉列表
func (s *MetricService) SelectMetric(ctx context.Context, req *datasourceapi.ListMetricRequest) (*datasourceapi.SelectMetricReply, error) {
	params := builder.NewParamsBuild(ctx).MetricModuleBuilder().WithListMetricRequest(req).ToBo()
	list, err := s.metricBiz.ListMetric(ctx, params)
	if err != nil {
		return nil, err
	}

	return &datasourceapi.SelectMetricReply{
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(params.Page),
		List:       builder.NewParamsBuild(ctx).MetricModuleBuilder().DoMetricBuilder().ToSelects(list),
	}, nil
}

// DeleteMetric 删除指标
func (s *MetricService) DeleteMetric(ctx context.Context, req *datasourceapi.DeleteMetricRequest) (*datasourceapi.DeleteMetricReply, error) {
	if err := s.metricBiz.DeleteMetricByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &datasourceapi.DeleteMetricReply{}, nil
}

// SyncMetric 同步指标
func (s *MetricService) SyncMetric(ctx context.Context, req *datasourceapi.SyncMetricRequest) (*datasourceapi.SyncMetricReply, error) {
	// 创建指标
	metricInfo := req.GetMetrics()
	createMetric := &bo.CreateMetricParams{
		Metric: &bo.MetricBo{
			Name: metricInfo.GetName(),
			Help: metricInfo.GetHelp(),
			Type: vobj.MetricType(metricInfo.GetType()),
			Unit: metricInfo.GetUnit(),
			Labels: types.SliceTo(metricInfo.GetLabels(), func(item *admin.MetricLabelItem) *bo.MetricLabel {
				return &bo.MetricLabel{
					Name:   item.GetName(),
					Values: item.GetValues(),
				}
			}),
		},
		Done:         req.GetDone(),
		DatasourceID: req.GetDatasourceId(),
		TeamID:       req.GetTeamId(),
	}
	if err := s.metricBiz.CreateMetric(ctx, createMetric); err != nil {
		return nil, err
	}
	return nil, nil
}
