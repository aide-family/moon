// Package cron provides cron-based background servers for Jade Tree.
package cron

import (
	"context"

	mcron "github.com/aide-family/magicbox/server/cron"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/jade_tree/internal/biz"
	"github.com/aide-family/jade_tree/internal/conf"
)

var _ transport.Server = (*MachineInfoReporterServer)(nil)

func NewMachineInfoReporterServer(bc *conf.Bootstrap, machineInfoBiz *biz.MachineInfo, helper *klog.Helper) *MachineInfoReporterServer {
	reportJob := mcron.WrapCronJobWithMetrics(NewMachineInfoReporterJob(bc, machineInfoBiz, helper))
	collectSelfJob := mcron.WrapCronJobWithMetrics(NewCollectSelfJob(bc, machineInfoBiz, helper))
	opts := []mcron.Option{
		mcron.WithCronJobs(reportJob, collectSelfJob),
	}
	return &MachineInfoReporterServer{
		Server: mcron.New("jade-tree-machine-info-reporter", helper.Logger(), opts...),
		helper: helper,
	}
}

type MachineInfoReporterServer struct {
	*mcron.Server
	helper *klog.Helper
}

func (s *MachineInfoReporterServer) Start(ctx context.Context) error {
	s.helper.WithContext(ctx).Infow("msg", "[MachineInfoReporter] started")
	return s.Server.Start(ctx)
}

func (s *MachineInfoReporterServer) Stop(ctx context.Context) error {
	defer s.helper.WithContext(ctx).Infow("msg", "[MachineInfoReporter] stopped")
	return s.Server.Stop(ctx)
}
