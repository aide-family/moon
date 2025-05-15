package biz

import (
	"context"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/houyi/internal/biz/do"
	"github.com/aide-family/moon/cmd/houyi/internal/biz/event"
	"github.com/aide-family/moon/cmd/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/houyi/internal/conf"
	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/plugin/server"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
)

func NewMetric(
	bc *conf.Bootstrap,
	judgeRepo repository.Judge,
	alertRepo repository.Alert,
	metricInitRepo repository.MetricInit,
	configRepo repository.Config,
	eventBusRepo repository.EventBus,
	cacheRepo repository.Cache,
	callbackRepo repository.Callback,
	logger log.Logger,
) *Metric {
	evaluateConf := bc.GetEvaluate()
	syncConfig := bc.GetConfig()
	return &Metric{
		helper:           log.NewHelper(log.With(logger, "module", "biz.metric")),
		judgeRepo:        judgeRepo,
		alertRepo:        alertRepo,
		metricInitRepo:   metricInitRepo,
		configRepo:       configRepo,
		eventBusRepo:     eventBusRepo,
		cacheRepo:        cacheRepo,
		callbackRepo:     callbackRepo,
		evaluateInterval: evaluateConf.GetInterval().AsDuration(),
		evaluateTimeout:  evaluateConf.GetTimeout().AsDuration(),
		syncInterval:     syncConfig.GetSyncInterval().AsDuration(),
		syncTimeout:      syncConfig.GetSyncTimeout().AsDuration(),
	}
}

type Metric struct {
	helper *log.Helper

	judgeRepo        repository.Judge
	alertRepo        repository.Alert
	metricInitRepo   repository.MetricInit
	configRepo       repository.Config
	eventBusRepo     repository.EventBus
	cacheRepo        repository.Cache
	callbackRepo     repository.Callback
	evaluateInterval time.Duration
	evaluateTimeout  time.Duration
	syncInterval     time.Duration
	syncTimeout      time.Duration
}

func (m *Metric) Loads() []*server.TickTask {
	return []*server.TickTask{
		{
			Fn:        m.syncMetricRuleConfigs,
			Name:      "syncMetricRuleConfigs",
			Timeout:   m.syncTimeout,
			Interval:  m.syncInterval,
			Immediate: true,
		},
	}
}

func (m *Metric) syncMetricRuleConfigs(ctx context.Context, isStop bool) error {
	if isStop {
		return nil
	}
	metricRules, err := m.configRepo.GetMetricRules(ctx)
	if err != nil {
		m.helper.WithContext(ctx).Errorw("method", "syncMetricRuleConfigs", "err", err)
		return err
	}

	return m.syncMetricJob(ctx, metricRules...)
}

func (m *Metric) syncMetricJob(ctx context.Context, rules ...bo.MetricRule) error {
	inStrategyJobEventBus := m.eventBusRepo.InStrategyJobEventBus()
	for _, rule := range rules {
		rule.Renovate()
		strategyJob, err := m.newStrategyJob(ctx, rule)
		if err != nil {
			m.helper.WithContext(ctx).Warnw("msg", "new strategy job error", "err", err)
			continue
		}
		inStrategyJobEventBus <- strategyJob
	}

	m.helper.WithContext(ctx).Debug("save metric rules success")
	return nil
}

func (m *Metric) SaveMetricRules(ctx context.Context, rules ...bo.MetricRule) error {
	if len(rules) == 0 {
		return nil
	}

	if err := m.configRepo.SetMetricRules(ctx, rules...); err != nil {
		m.helper.WithContext(ctx).Errorw("msg", "save metric rules error", "err", err)
		return err
	}

	return m.syncMetricJob(ctx, rules...)
}

func (m *Metric) newStrategyJob(_ context.Context, metric bo.MetricRule) (bo.StrategyJob, error) {
	opts := []event.StrategyMetricJobOption{
		event.WithStrategyMetricJobHelper(m.helper.Logger()),
		event.WithStrategyMetricJobMetric(metric.UniqueKey(), metric.GetEnable()),
		event.WithStrategyMetricJobConfigRepo(m.configRepo),
		event.WithStrategyMetricJobJudgeRepo(m.judgeRepo),
		event.WithStrategyMetricJobAlertRepo(m.alertRepo),
		event.WithStrategyMetricJobMetricInitRepo(m.metricInitRepo),
		event.WithStrategyMetricJobSpec(m.evaluateInterval),
		event.WithStrategyMetricJobTimeout(m.evaluateTimeout),
		event.WithStrategyMetricJobEventBusRepo(m.eventBusRepo),
		event.WithStrategyMetricJobCacheRepo(m.cacheRepo),
	}
	return event.NewStrategyMetricJob(metric.UniqueKey(), opts...)
}

func (m *Metric) SyncMetricMetadata(ctx context.Context, req *bo.SyncMetricMetadataRequest) error {
	metricInstance, err := m.metricInitRepo.Init(req.Item)
	if err != nil {
		m.helper.WithContext(ctx).Errorw("msg", "sync metric metadata error", "err", err)
		return err
	}

	ts := timex.Now()
	metadataChan, err := metricInstance.Metadata(ctx)
	if err != nil {
		m.helper.WithContext(ctx).Errorw("msg", "sync metric metadata error", "err", err)
		return err
	}

	total := 0
	for metadata := range metadataChan {
		total += len(metadata)
		metadataItems := slices.Map(metadata, func(v *do.MetricItem) *common.MetricItem {
			labels := make(map[string]string)
			for k, v := range v.Labels {
				labels[k] = strings.Join(v, ",")
			}
			return &common.MetricItem{
				Name:   v.Name,
				Help:   v.Help,
				Type:   v.Type,
				Labels: labels,
				Unit:   v.Unit,
			}
		})
		params := &palace.SyncMetadataRequest{
			Items:        metadataItems,
			OperatorId:   req.OperatorId,
			TeamId:       req.Item.GetTeamId(),
			DatasourceId: req.Item.GetId(),
		}
		if err := m.callbackRepo.SyncMetadata(ctx, params); err != nil {
			m.helper.WithContext(ctx).Errorw("msg", "sync metric metadata error", "err", err)
			return err
		}
	}
	m.helper.WithContext(ctx).Debugf("total metric: %d, cost: %s", total, time.Since(ts))

	params := &palace.SyncMetadataRequest{
		IsDone:       true,
		OperatorId:   req.OperatorId,
		TeamId:       req.Item.GetTeamId(),
		DatasourceId: req.Item.GetId(),
	}
	if err := m.callbackRepo.SyncMetadata(ctx, params); err != nil {
		m.helper.WithContext(ctx).Errorw("msg", "sync metric metadata error", "err", err)
		return err
	}
	return nil
}

func (m *Metric) QueryMetricDatasource(ctx context.Context, req *bo.MetricDatasourceQueryRequest) (*bo.MetricDatasourceQueryReply, error) {
	metricInstance, err := m.metricInitRepo.Init(req.Datasource)
	if err != nil {
		m.helper.WithContext(ctx).Errorw("msg", "query metric datasource error", "err", err)
		return nil, err
	}

	if req.EndTime > req.StartTime && req.EndTime > 0 {
		queryRangeRequest := &bo.MetricRangeQueryRequest{
			Expr:      req.Expr,
			StartTime: time.Unix(req.StartTime, 0),
			EndTime:   time.Unix(req.EndTime, 0),
		}
		queryResponse, err := metricInstance.QueryRange(ctx, queryRangeRequest)
		if err != nil {
			m.helper.WithContext(ctx).Errorw("msg", "query metric datasource error", "err", err)
			return nil, err
		}
		return NewMetricDatasourceQueryReply(WithMetricDatasourceQueryRangeReply(queryResponse)), nil
	}

	queryRequest := &bo.MetricQueryRequest{
		Expr: req.Expr,
		Time: time.Unix(req.Time, 0),
	}
	queryResponse, err := metricInstance.Query(ctx, queryRequest)
	if err != nil {
		m.helper.WithContext(ctx).Errorw("msg", "query metric datasource error", "err", err)
		return nil, err
	}
	return NewMetricDatasourceQueryReply(WithMetricDatasourceQueryReplyResults(queryResponse)), nil
}
