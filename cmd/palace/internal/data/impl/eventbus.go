package impl

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/queue/ringbuffer"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ repository.EventBus = (*eventBusImpl)(nil)

func NewEventBus(d *data.Data) repository.EventBus {
	eventBus := &eventBusImpl{
		dataChangeEventBus: d.GetDataChangeEventBus(),
		cacheBuf:           d.GetRingBuffer(),
	}
	eventBus.cacheBuf.RegisterOnTrigger(eventBus.syncRequest)
	return eventBus
}

type eventBusImpl struct {
	dataChangeEventBus chan *bo.SyncRequest
	cacheBuf           *ringbuffer.RingBuffer[*bo.SyncRequest]
}

// PublishDataChangeEvent implements repository.EventBus.
func (e *eventBusImpl) PublishDataChangeEvent(eventType vobj.ChangedType, teamID uint32, ids ...uint32) {
	ids = slices.MapFilter(ids, func(id uint32) (uint32, bool) {
		return id, id > 0
	})
	if len(ids) == 0 || teamID == 0 {
		return
	}

	e.cacheBuf.Add(&bo.SyncRequest{
		Rows: bo.ChangedRows{
			teamID: ids,
		},
		Type: eventType,
	})
}

// SubscribeDataChangeEvent implements repository.EventBus.
func (e *eventBusImpl) SubscribeDataChangeEvent() <-chan *bo.SyncRequest {
	return e.dataChangeEventBus
}

func (e *eventBusImpl) syncRequest(items []*bo.SyncRequest) {
	pushedItem := make(map[vobj.ChangedType]bo.ChangedRows)
	for _, item := range items {
		if _, ok := pushedItem[item.Type]; !ok {
			pushedItem[item.Type] = make(bo.ChangedRows, len(item.Rows))
		}
		for teamID, ids := range item.Rows {
			if _, ok := pushedItem[item.Type][teamID]; !ok {
				pushedItem[item.Type][teamID] = make([]uint32, 0, len(ids))
			}
			pushedItem[item.Type][teamID] = slices.Unique(append(pushedItem[item.Type][teamID], ids...))
		}
	}
	for eventType, rows := range pushedItem {
		e.dataChangeEventBus <- &bo.SyncRequest{
			Rows: rows,
			Type: eventType,
		}
	}
}
