package impl

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/data"
)

var _ repository.EventBus = (*eventBusImpl)(nil)

func NewEventBus(d *data.Data) repository.EventBus {
	return &eventBusImpl{
		dataChangeEventBus: d.GetDataChangeEventBus(),
	}
}

type eventBusImpl struct {
	dataChangeEventBus chan *bo.SyncRequest
}

// PublishDataChangeEvent implements repository.EventBus.
func (e *eventBusImpl) PublishDataChangeEvent(event *bo.SyncRequest) {
	e.dataChangeEventBus <- event
}

// SubscribeDataChangeEvent implements repository.EventBus.
func (e *eventBusImpl) SubscribeDataChangeEvent() <-chan *bo.SyncRequest {
	return e.dataChangeEventBus
}
