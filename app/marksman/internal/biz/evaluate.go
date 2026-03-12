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
)

func NewEvaluateBiz(
	namespaceRepo repository.Namespace,
	strategyMetricRepo repository.StrategyMetric,
	jobChannelRepo repository.JobChannel,
) *Evaluate {
	eg := new(errgroup.Group)
	eg.SetLimit(10)
	eva := &Evaluate{
		namespaceRepo:      namespaceRepo,
		strategyMetricRepo: strategyMetricRepo,
		jobChannelRepo:     jobChannelRepo,
		eg:                 eg,
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
}

func (e *Evaluate) Start() {
	e.eg.Go(func() error {
		time.Sleep(10 * time.Second)
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
	ctx := context.Background()
	for {
		namespaces, lastUID, hasMore, err := e.getNamespaces(ctx, req)
		if err != nil {
			klog.Errorw("msg", "select namespace failed", "error", err)
			break
		}
		for _, namespace := range namespaces {
			e.localStrategyMetricsByNamespace(ctx, eg, snowflake.ID(namespace.Value))
		}
		if !hasMore {
			break
		}
		req.LastUID = lastUID
	}
}

func (e *Evaluate) getNamespaces(ctx context.Context, req *goddessv1.SelectNamespaceRequest) ([]*goddessv1.NamespaceItemSelect, int64, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	namespacesReply, err := e.namespaceRepo.SelectNamespace(ctx, req)
	if err != nil {
		return nil, 0, false, err
	}
	return namespacesReply.Items, namespacesReply.LastUID, namespacesReply.HasMore, nil
}

func (e *Evaluate) localStrategyMetricsByNamespace(ctx context.Context, eg *errgroup.Group, namespaceUID snowflake.ID) {
	ctx = contextx.WithNamespace(ctx, namespaceUID)
	eg.Go(func() error {
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
