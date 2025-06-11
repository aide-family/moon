package impl

import (
	"sync"
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
)

var _ repository.EventBus = (*eventBusImpl)(nil)

func NewEventBus(d *data.Data) repository.EventBus {
	eventBus := &eventBusImpl{
		dataChangeEventBus: d.GetDataChangeEventBus(),
		rows:               make(map[vobj.ChangedType]map[uint32][]uint32),
		cacheBuf:           make(chan *bo.SyncRequest, 10000),
	}
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			eventBus.publishDataChangeEvent()
		}
	}()
	go func() {
		for event := range eventBus.cacheBuf {
			eventBus.dataChangeEventBus <- event
		}
	}()
	return eventBus
}

type eventBusImpl struct {
	lock               sync.Mutex
	dataChangeEventBus chan *bo.SyncRequest
	cacheBuf           chan *bo.SyncRequest
	rows               map[vobj.ChangedType]map[uint32][]uint32
}

// PublishDataChangeEvent implements repository.EventBus.
func (e *eventBusImpl) PublishDataChangeEvent(eventType vobj.ChangedType, teamID uint32, id uint32) {
	e.lock.Lock()
	defer e.lock.Unlock()
	rows, ok := e.rows[eventType]
	if !ok {
		rows = make(map[uint32][]uint32)
	}
	rows[teamID] = append(rows[teamID], id)
	e.rows[eventType] = rows
}

// SubscribeDataChangeEvent implements repository.EventBus.
func (e *eventBusImpl) SubscribeDataChangeEvent() <-chan *bo.SyncRequest {
	return e.dataChangeEventBus
}

func (e *eventBusImpl) publishDataChangeEvent() {
	e.lock.Lock()
	defer e.lock.Unlock()
	for eventType, rows := range e.rows {
		changedRows := make(bo.ChangedRows)
		for teamID, ids := range rows {
			if len(ids) > 0 {
				changedRows[teamID] = ids
			}
		}
		if len(changedRows) > 0 {
			e.cacheBuf <- &bo.SyncRequest{
				Rows: changedRows,
				Type: eventType,
			}
		}
	}
	e.rows = make(map[vobj.ChangedType]map[uint32][]uint32)
}
