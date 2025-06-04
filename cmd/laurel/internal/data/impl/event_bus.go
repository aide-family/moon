package impl

import (
	"github.com/aide-family/moon/cmd/laurel/internal/biz/repository"
	"github.com/aide-family/moon/cmd/laurel/internal/data"
)

func NewEventBusRepo(d *data.Data) repository.EventBus {
	return &eventBusRepoImpl{Data: d}
}

type eventBusRepoImpl struct {
	*data.Data
}
