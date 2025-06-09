package service

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
)

func NewLoadService(
	userBiz *biz.UserBiz,
	teamBiz *biz.Team,
	menuBiz *biz.Menu,
	eventBus *biz.EventBus,
) *LoadService {
	return &LoadService{
		userBiz:  userBiz,
		teamBiz:  teamBiz,
		menuBiz:  menuBiz,
		eventBus: eventBus,
	}
}

type LoadService struct {
	userBiz  *biz.UserBiz
	teamBiz  *biz.Team
	menuBiz  *biz.Menu
	eventBus *biz.EventBus
}

func (s *LoadService) LoadJobs() []cron_server.CronJob {
	userJobs := s.userBiz.Jobs()
	teamJobs := s.teamBiz.Jobs()
	menuJobs := s.menuBiz.Jobs()
	return append(append(userJobs, teamJobs...), menuJobs...)
}

func (s *LoadService) SubscribeDataChangeEvent() <-chan *bo.SyncRequest {
	return s.eventBus.SubscribeDataChangeEvent()
}
