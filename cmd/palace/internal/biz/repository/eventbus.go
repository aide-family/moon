package repository

import "github.com/aide-family/moon/cmd/palace/internal/biz/bo"

type EventBus interface {
	PublishDataChangeEvent(event *bo.SyncRequest)
	SubscribeDataChangeEvent() <-chan *bo.SyncRequest
}
