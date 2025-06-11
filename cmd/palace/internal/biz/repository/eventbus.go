package repository

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type EventBus interface {
	PublishDataChangeEvent(eventType vobj.ChangedType, teamID uint32, id uint32)
	SubscribeDataChangeEvent() <-chan *bo.SyncRequest
}
