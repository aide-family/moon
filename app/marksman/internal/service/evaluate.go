package service

import "github.com/aide-family/magicbox/server/cron"

func NewEvaluateService() *EvaluateService {
	return &EvaluateService{}
}

type EvaluateService struct{}

func (s *EvaluateService) GetMetricAppendJobChannel() <-chan cron.CronJob {
	return make(<-chan cron.CronJob, 100)
}

func (s *EvaluateService) GetMetricRemoveJobChannel() <-chan string {
	return make(<-chan string, 100)
}
