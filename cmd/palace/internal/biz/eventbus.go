package biz

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
)

func NewEventBus(eventBusRepo repository.EventBus) *EventBus {
	return &EventBus{
		eventBusRepo: eventBusRepo,
	}
}

type EventBus struct {
	eventBusRepo repository.EventBus
}

func (e *EventBus) SubscribeDataChangeEvent() <-chan *bo.SyncRequest {
	return e.eventBusRepo.SubscribeDataChangeEvent()
}
