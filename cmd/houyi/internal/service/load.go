package service

import (
	"github.com/aide-family/moon/cmd/houyi/internal/biz"
	"github.com/aide-family/moon/pkg/plugin/server/ticker_server"
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

func (s *LoadService) Loads() []*ticker_server.TickTask {
	return append(s.metricBiz.Loads(), s.alertBiz.Loads()...)
}
