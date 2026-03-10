// Package cron is the cron server for the marksman.
package cron

import (
	"github.com/aide-family/magicbox/server/cron"
	"github.com/aide-family/marksman/internal/service"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

var _ transport.Server = (*MetricCronServer)(nil)

func NewMetricCronServer(evaluateService *service.EvaluateService, helper *klog.Helper) *MetricCronServer {
	name := "marksman-cron-metric"
	opts := []cron.Option{
		cron.WithCronJobChannel(evaluateService.GetMetricAppendJobChannel()),
		cron.WithRemoveJobChannel(evaluateService.GetMetricRemoveJobChannel()),
	}
	return &MetricCronServer{
		Server: cron.New(name, helper.Logger(), opts...),
	}
}

type MetricCronServer struct {
	*cron.Server
}
