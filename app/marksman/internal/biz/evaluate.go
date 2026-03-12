package biz

import (
	"context"
	"time"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/server/cron"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"

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
	jobChannelRepo repository.JobChannel,
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
		namespaceRepo:      namespaceRepo,
		strategyMetricRepo: strategyMetricRepo,
		jobChannelRepo:     jobChannelRepo,
		eg:                 eg,
		startupDelay:       startupDelay,
		queryTimeout:       queryTimeout,
	}
	jobChannelRepo.AppendClose(eva.Stop)
	eva.Start()
	return eva
}

type Evaluate struct {
	namespaceRepo      repository.Namespace
	strategyMetricRepo repository.StrategyMetric
	jobChannelRepo     repository.JobChannel
	eg                 *errgroup.Group
	startupDelay       time.Duration
	queryTimeout       time.Duration
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

func (e *Evaluate) GetMetricAppendJobChannel() <-chan cron.CronJob {
	return e.jobChannelRepo.GetMetricAppendJobChannel()
}

func (e *Evaluate) GetMetricRemoveJobChannel() <-chan string {
	return e.jobChannelRepo.GetMetricRemoveJobChannel()
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
			e.jobChannelRepo.AppendMetricJob(evaluator.NewMetricEvaluator(strategy))
		}
		return nil
	})
}
