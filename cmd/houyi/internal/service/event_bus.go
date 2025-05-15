package service

import (
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/houyi/internal/biz"
	"github.com/aide-family/moon/cmd/houyi/internal/biz/bo"
)

func NewEventBusService(
	eventBus *biz.EventBus,
	logger log.Logger,
) *EventBusService {
	return &EventBusService{
		helper:   log.NewHelper(log.With(logger, "module", "service.event-bus")),
		eventBus: eventBus,
	}
}

type EventBusService struct {
	helper   *log.Helper
	eventBus *biz.EventBus
}

func (s *EventBusService) OutStrategyJobEventBus() <-chan bo.StrategyJob {
	return s.eventBus.OutStrategyJobEventBus()
}

func (s *EventBusService) OutAlertJobEventBus() <-chan bo.AlertJob {
	return s.eventBus.OutAlertJobEventBus()
}

func (s *EventBusService) OutAlertEventBus() <-chan bo.Alert {
	return s.eventBus.OutAlertEventBus()
}
