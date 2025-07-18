package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/houyi/internal/biz/do"
	"github.com/aide-family/moon/cmd/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/houyi/internal/data"
	"github.com/aide-family/moon/pkg/api/houyi/common"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/datasource"
	"github.com/aide-family/moon/pkg/plugin/datasource/prometheus"
	"github.com/aide-family/moon/pkg/plugin/datasource/victoria"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewMetricRepo(d *data.Data, logger log.Logger) repository.MetricInit {
	return &metricImpl{
		Data: d,
		help: log.NewHelper(log.With(logger, "module", "data.repo.metric")),
	}
}

type metricImpl struct {
	*data.Data
	help *log.Helper
}

type metricInstance struct {
	metric datasource.Metric
	helper *log.Helper
}

func (m *metricImpl) Init(config bo.MetricDatasourceConfig) (repository.Metric, error) {
	if validate.IsNil(config) {
		return nil, merr.ErrorInvalidArgument("metric datasource config is nil")
	}

	var (
		metricDatasource datasource.Metric
		ok               bool
	)

	metricDatasource, ok = m.GetMetricDatasource(config.UniqueKey())
	switch config.GetDriver() {
	case common.MetricDatasourceDriver_PROMETHEUS:
		if !ok {
			metricDatasource = prometheus.New(config, m.help.Logger())
		}
	case common.MetricDatasourceDriver_VICTORIAMETRICS:
		if !ok {
			metricDatasource = victoria.New(config, m.help.Logger())
		}
	default:
		return nil, merr.ErrorParams("invalid metric datasource driver: %s", config.GetDriver())
	}
	return &metricInstance{
		metric: metricDatasource,
		helper: log.NewHelper(log.With(m.help.Logger(), "module", "data.repo.metric.instance")),
	}, nil
}

func (m *metricInstance) Query(ctx context.Context, req *bo.MetricQueryRequest) ([]*do.MetricQueryReply, error) {
	queryParams := &datasource.MetricQueryRequest{
		Expr: req.Expr,
		Time: req.Time.Unix(),
	}
	metricQueryResponse, err := m.metric.Query(ctx, queryParams)
	if err != nil {
		m.helper.WithContext(ctx).Warnw("msg", "query metric failed", "err", err)
		return nil, err
	}
	if validate.IsNil(metricQueryResponse) || validate.IsNil(metricQueryResponse.Data) {
		return nil, nil
	}
	list := make([]*do.MetricQueryReply, 0, len(metricQueryResponse.Data.Result))
	for _, result := range metricQueryResponse.Data.Result {
		queryValue := result.GetMetricQueryValue()
		item := &do.MetricQueryReply{
			Labels: result.Metric,
			Value: &do.MetricQueryValue{
				Value:     queryValue.Value,
				Timestamp: int64(queryValue.Timestamp),
			},
			ResultType: string(metricQueryResponse.Data.ResultType),
		}
		list = append(list, item)
	}
	return list, nil
}

func (m *metricInstance) QueryRange(ctx context.Context, req *bo.MetricRangeQueryRequest) ([]*do.MetricQueryRangeReply, error) {
	// Calculate resolution
	step := req.GetOptimalStep(m.metric.GetScrapeInterval())
	queryParams := &datasource.MetricQueryRequest{
		Expr:      req.Expr,
		StartTime: req.StartTime.Unix(),
		EndTime:   req.EndTime.Unix(),
		Step:      uint32(step.Seconds()),
	}
	metricQueryResponse, err := m.metric.Query(ctx, queryParams)
	if err != nil {
		m.helper.Warnw("msg", "query metric range failed", "err", err)
		return nil, err
	}
	if validate.IsNil(metricQueryResponse) || validate.IsNil(metricQueryResponse.Data) {
		return nil, nil
	}
	list := make([]*do.MetricQueryRangeReply, 0, len(metricQueryResponse.Data.Result))
	for _, result := range metricQueryResponse.Data.Result {
		queryValues := result.GetMetricQueryValues()
		item := &do.MetricQueryRangeReply{
			Labels:     result.Metric,
			Values:     make([]*do.MetricQueryValue, 0, len(queryValues)),
			ResultType: string(metricQueryResponse.Data.ResultType),
		}
		for _, queryValue := range queryValues {
			item.Values = append(item.Values, &do.MetricQueryValue{
				Value:     queryValue.Value,
				Timestamp: int64(queryValue.Timestamp),
			})
		}
		list = append(list, item)
	}
	return list, nil
}

func (m *metricInstance) Metadata(ctx context.Context) (<-chan []*do.MetricItem, error) {
	metricMetadata, err := m.metric.Metadata(ctx)
	if err != nil {
		m.helper.Warnw("msg", "get metric metadata failed", "err", err)
		return nil, err
	}
	ch := make(chan []*do.MetricItem)
	safety.Go("metricInstance.Metadata", func() {
		defer close(ch)
		for metadata := range metricMetadata {
			syncList := make([]*do.MetricItem, 0, len(metadata.Metric))
			for _, metricMetadataItem := range metadata.Metric {
				item := &do.MetricItem{
					Name:   metricMetadataItem.Name,
					Help:   metricMetadataItem.Help,
					Type:   metricMetadataItem.Type,
					Labels: metricMetadataItem.Labels,
					Unit:   metricMetadataItem.Unit,
				}
				syncList = append(syncList, item)
			}
			ch <- syncList
		}
	}, m.helper.Logger())
	return ch, nil
}
