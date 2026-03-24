package biz

import (
	"context"
	"sync"
	"time"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/server/cron"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/evaluator"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
)

const (
	defaultConcurrencyLimit = 10
	defaultStartupDelay     = 10 * time.Second
	defaultQueryTimeout     = 10 * time.Second
)

func NewEvaluateBiz(
	bc *conf.Bootstrap,
	namespaceRepo repository.Namespace,
	strategyMetricRepo repository.StrategyMetric,
	evaluateJobChannelRepo repository.EvaluateJobChannel,
	metricQuerier repository.MetricDatasourceQuerier,
	alertEventChannel repository.AlertEventChannel,
) *Evaluate {
	limit := defaultConcurrencyLimit
	startupDelay := defaultStartupDelay
	queryTimeout := defaultQueryTimeout
	if cfg := bc.GetEvaluateConfig(); cfg != nil {
		if cfg.GetConcurrencyLimit() > 0 {
			limit = int(cfg.GetConcurrencyLimit())
		}
		if cfg.GetStartupDelay() != nil {
			startupDelay = cfg.GetStartupDelay().AsDuration()
		}
		if cfg.GetQueryTimeout() != nil {
			queryTimeout = cfg.GetQueryTimeout().AsDuration()
		}
	}
	eg := new(errgroup.Group)
	eg.SetLimit(limit)
	eva := &Evaluate{
		namespaceRepo:          namespaceRepo,
		strategyMetricRepo:     strategyMetricRepo,
		evaluateJobChannelRepo: evaluateJobChannelRepo,
		metricQuerier:          metricQuerier,
		alertEventChannel:      alertEventChannel,
		eg:                     eg,
		startupDelay:           startupDelay,
		queryTimeout:           queryTimeout,
		jobState:               make(map[string]evaluateJobMeta),
	}
	evaluateJobChannelRepo.AppendClose(eva.Stop)
	eva.Start()
	return eva
}

type Evaluate struct {
	namespaceRepo          repository.Namespace
	strategyMetricRepo     repository.StrategyMetric
	evaluateJobChannelRepo repository.EvaluateJobChannel
	metricQuerier          repository.MetricDatasourceQuerier
	alertEventChannel      repository.AlertEventChannel
	eg                     *errgroup.Group
	startupDelay           time.Duration
	queryTimeout           time.Duration
	jobStateMu             sync.RWMutex
	jobState               map[string]evaluateJobMeta
}

type evaluateJobMeta struct {
	namespaceUID     snowflake.ID
	datasourceUID    snowflake.ID
	strategyGroupUID snowflake.ID
	strategyUID      snowflake.ID
	levelUID         snowflake.ID
}

func (e *Evaluate) Start() {
	e.eg.Go(func() error {
		time.Sleep(e.startupDelay)
		e.loadAllStrategyMetrics(e.eg)
		return nil
	})
}

func (e *Evaluate) Stop() error {
	return e.eg.Wait()
}

func (e *Evaluate) GetEvaluateJobAppendChannel() <-chan cron.CronJob {
	return e.evaluateJobChannelRepo.GetEvaluateJobAppendChannel()
}

func (e *Evaluate) GetEvaluateJobRemoveChannel() <-chan string {
	return e.evaluateJobChannelRepo.GetEvaluateJobRemoveChannel()
}

func (e *Evaluate) loadAllStrategyMetrics(eg *errgroup.Group) {
	req := &goddessv1.SelectNamespaceRequest{
		Limit:   500,
		Status:  enum.GlobalStatus_ENABLED,
		LastUID: 0,
	}

	for {
		namespaces, lastUID, hasMore, err := e.getNamespaces(req)
		if err != nil {
			klog.Errorw("msg", "select namespace failed", "error", err)
			break
		}
		for _, namespace := range namespaces {
			e.localStrategyMetricsByNamespace(eg, snowflake.ID(namespace.Value))
		}
		if !hasMore {
			break
		}
		req.LastUID = lastUID
	}
}

func (e *Evaluate) getNamespaces(req *goddessv1.SelectNamespaceRequest) ([]*goddessv1.NamespaceItemSelect, int64, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.queryTimeout)
	defer cancel()
	namespacesReply, err := e.namespaceRepo.SelectNamespace(ctx, req)
	if err != nil {
		return nil, 0, false, err
	}
	return namespacesReply.Items, namespacesReply.LastUID, namespacesReply.HasMore, nil
}

func (e *Evaluate) localStrategyMetricsByNamespace(eg *errgroup.Group, namespaceUID snowflake.ID) {
	eg.Go(func() error {
		ctx := contextx.WithNamespace(context.Background(), namespaceUID)
		ctx, cancel := context.WithTimeout(ctx, e.queryTimeout)
		defer cancel()
		strategies, err := e.strategyMetricRepo.GetEvaluateMetricStrategies(ctx)
		if err != nil {
			klog.Errorw("msg", "get evaluate metric strategies failed", "error", err, "namespaceUID", namespaceUID)
			return err
		}

		for _, strategy := range strategies {
			e.appendEvaluateStrategyJob(strategy)
		}
		return nil
	})
}

func (e *Evaluate) SyncByDatasourceUID(ctx context.Context, datasourceUID snowflake.ID) {
	e.syncByFilter(ctx, func(meta evaluateJobMeta) bool {
		return meta.datasourceUID == datasourceUID
	}, func(strategy *bo.EvaluateMetricStrategyBo) bool {
		return strategy.GetDatasourceUID() == datasourceUID
	})
}

func (e *Evaluate) RemoveByDatasourceUID(ctx context.Context, datasourceUID snowflake.ID) {
	e.removeByFilter(ctx, func(meta evaluateJobMeta) bool {
		return meta.datasourceUID == datasourceUID
	})
}

func (e *Evaluate) SyncByStrategyGroupUID(ctx context.Context, strategyGroupUID snowflake.ID) {
	e.syncByFilter(ctx, func(meta evaluateJobMeta) bool {
		return meta.strategyGroupUID == strategyGroupUID
	}, func(strategy *bo.EvaluateMetricStrategyBo) bool {
		return strategy.GetStrategyGroupUID() == strategyGroupUID
	})
}

func (e *Evaluate) RemoveByStrategyGroupUID(ctx context.Context, strategyGroupUID snowflake.ID) {
	e.removeByFilter(ctx, func(meta evaluateJobMeta) bool {
		return meta.strategyGroupUID == strategyGroupUID
	})
}

func (e *Evaluate) SyncByStrategyUID(ctx context.Context, strategyUID snowflake.ID) {
	e.syncByFilter(ctx, func(meta evaluateJobMeta) bool {
		return meta.strategyUID == strategyUID
	}, func(strategy *bo.EvaluateMetricStrategyBo) bool {
		return strategy.GetStrategyUID() == strategyUID
	})
}

func (e *Evaluate) RemoveByStrategyUID(ctx context.Context, strategyUID snowflake.ID) {
	e.removeByFilter(ctx, func(meta evaluateJobMeta) bool {
		return meta.strategyUID == strategyUID
	})
}

func (e *Evaluate) SyncByStrategyLevelUID(ctx context.Context, strategyUID snowflake.ID, levelUID snowflake.ID) {
	e.syncByFilter(ctx, func(meta evaluateJobMeta) bool {
		return meta.strategyUID == strategyUID && meta.levelUID == levelUID
	}, func(strategy *bo.EvaluateMetricStrategyBo) bool {
		return strategy.GetStrategyUID() == strategyUID && strategy.GetLevelUID() == levelUID
	})
}

func (e *Evaluate) RemoveByStrategyLevelUID(ctx context.Context, strategyUID snowflake.ID, levelUID snowflake.ID) {
	e.removeByFilter(ctx, func(meta evaluateJobMeta) bool {
		return meta.strategyUID == strategyUID && meta.levelUID == levelUID
	})
}

func (e *Evaluate) SyncByLevelUID(ctx context.Context, levelUID snowflake.ID) {
	e.syncByFilter(ctx, func(meta evaluateJobMeta) bool {
		return meta.levelUID == levelUID
	}, func(strategy *bo.EvaluateMetricStrategyBo) bool {
		return strategy.GetLevelUID() == levelUID
	})
}

func (e *Evaluate) RemoveByLevelUID(ctx context.Context, levelUID snowflake.ID) {
	e.removeByFilter(ctx, func(meta evaluateJobMeta) bool {
		return meta.levelUID == levelUID
	})
}

func (e *Evaluate) syncByFilter(
	ctx context.Context,
	removeFilter func(meta evaluateJobMeta) bool,
	appendFilter func(strategy *bo.EvaluateMetricStrategyBo) bool,
) {
	e.removeByFilter(ctx, removeFilter)
	strategies, err := e.strategyMetricRepo.GetEvaluateMetricStrategies(ctx)
	if err != nil {
		klog.Errorw("msg", "get evaluate metric strategies failed", "error", err)
		return
	}
	for _, strategy := range strategies {
		if !appendFilter(strategy) {
			continue
		}
		e.appendEvaluateStrategyJob(strategy)
	}
}

func (e *Evaluate) removeByFilter(ctx context.Context, filter func(meta evaluateJobMeta) bool) {
	namespaceUID := contextx.GetNamespace(ctx)
	removeIndexes := e.collectJobIndexes(namespaceUID, filter)
	for _, index := range removeIndexes {
		e.removeEvaluateJob(index)
	}
}

func (e *Evaluate) collectJobIndexes(namespaceUID snowflake.ID, filter func(meta evaluateJobMeta) bool) []string {
	e.jobStateMu.RLock()
	defer e.jobStateMu.RUnlock()

	removeIndexes := make([]string, 0)
	for index, meta := range e.jobState {
		if meta.namespaceUID != namespaceUID {
			continue
		}
		if filter(meta) {
			removeIndexes = append(removeIndexes, index)
		}
	}
	return removeIndexes
}

func (e *Evaluate) appendEvaluateStrategyJob(strategy *bo.EvaluateMetricStrategyBo) {
	if strategy == nil {
		klog.Debugw("msg", "strategy is nil")
		return
	}
	index := strategy.BuildMetricEvaluatorIndex()
	e.jobStateMu.Lock()
	e.jobState[index] = evaluateJobMeta{
		namespaceUID:     strategy.GetNamespaceUID(),
		datasourceUID:    strategy.GetDatasourceUID(),
		strategyGroupUID: strategy.GetStrategyGroupUID(),
		strategyUID:      strategy.GetStrategyUID(),
		levelUID:         strategy.GetLevelUID(),
	}
	e.jobStateMu.Unlock()
	e.evaluateJobChannelRepo.AppendEvaluateJob(evaluator.NewMetricEvaluator(e.metricQuerier, e.alertEventChannel, strategy))
}

func (e *Evaluate) removeEvaluateJob(index string) {
	e.jobStateMu.Lock()
	delete(e.jobState, index)
	e.jobStateMu.Unlock()
	e.evaluateJobChannelRepo.RemoveEvaluateJob(index)
}
