package state

import (
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
)

var (
	messageStateProcessRegistry = safety.NewSyncMap(make(map[enum.MessageStatus]ProcessFunc))
	messageTaskStateRegistry    = safety.NewSyncMap(make(map[enum.MessageStatus]MessageTaskState))
)

func RegisterMessageTaskProcess(status enum.MessageStatus, processFunc ProcessFunc) {
	messageStateProcessRegistry.Set(status, processFunc)
}

func GetMessageTaskProcess(status enum.MessageStatus) (ProcessFunc, bool) {
	return messageStateProcessRegistry.Get(status)
}

func RegisterMessageTaskState(status enum.MessageStatus, state MessageTaskState) {
	messageTaskStateRegistry.Set(status, state)
}

func GetMessageTaskState(status enum.MessageStatus) (MessageTaskState, bool) {
	return messageTaskStateRegistry.Get(status)
}
