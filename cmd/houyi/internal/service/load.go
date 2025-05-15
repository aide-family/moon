package service

import (
	"github.com/moon-monitor/moon/cmd/houyi/internal/biz"
	"github.com/moon-monitor/moon/pkg/plugin/server"
)

func NewLoadService(
	alertBiz *biz.Alert,
	metricBiz *biz.Metric,
) *LoadService {
	return &LoadService{
		alertBiz:  alertBiz,
		metricBiz: metricBiz,
	}
}

type LoadService struct {
	alertBiz  *biz.Alert
	metricBiz *biz.Metric
}

func (s *LoadService) Loads() []*server.TickTask {
	return append(s.metricBiz.Loads(), s.alertBiz.Loads()...)
}
