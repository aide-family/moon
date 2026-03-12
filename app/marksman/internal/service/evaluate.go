package service

import (
	"github.com/aide-family/magicbox/server/cron"
	"github.com/aide-family/marksman/internal/biz"
)

func NewEvaluateService(evaluateBiz *biz.Evaluate) *EvaluateService {
	return &EvaluateService{evaluateBiz: evaluateBiz}
}

type EvaluateService struct {
	evaluateBiz *biz.Evaluate
}

func (s *EvaluateService) GetMetricAppendJobChannel() <-chan cron.CronJob {
	return s.evaluateBiz.GetMetricAppendJobChannel()
}

func (s *EvaluateService) GetMetricRemoveJobChannel() <-chan string {
	return s.evaluateBiz.GetMetricRemoveJobChannel()
}
