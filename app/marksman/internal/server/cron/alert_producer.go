// Package cron is the cron server for the marksman.
package cron

import (
	"github.com/aide-family/magicbox/server/cron"
	"github.com/aide-family/marksman/internal/service"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

var _ transport.Server = (*ProducerServer)(nil)

func NewProducerServer(evaluateService *service.EvaluateService, helper *klog.Helper) *ProducerServer {
	name := "marksman-cron-alert-producer"
	opts := []cron.Option{
		cron.WithCronJobChannel(evaluateService.GetEvaluateJobAppendChannel()),
		cron.WithRemoveJobChannel(evaluateService.GetEvaluateJobRemoveChannel()),
	}
	return &ProducerServer{
		Server: cron.New(name, helper.Logger(), opts...),
	}
}

type ProducerServer struct {
	*cron.Server
}
