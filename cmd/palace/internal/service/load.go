package service

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz"
	"github.com/moon-monitor/moon/pkg/plugin/server"
)

func NewLoadService(userBiz *biz.UserBiz, teamBiz *biz.Team) *LoadService {
	return &LoadService{
		userBiz: userBiz,
		teamBiz: teamBiz,
	}
}

type LoadService struct {
	userBiz *biz.UserBiz
	teamBiz *biz.Team
}

func (s *LoadService) LoadJobs() []server.CronJob {
	userJobs := s.userBiz.Jobs()
	return append(userJobs, s.teamBiz.Jobs()...)
}
