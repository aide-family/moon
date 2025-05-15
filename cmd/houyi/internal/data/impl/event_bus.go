package impl

import (
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/houyi/internal/data"
)

func NewEventBusRepo(d *data.Data, logger log.Logger) repository.EventBus {
	return &eventBusImpl{
		Data:   d,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.event-bus")),
	}
}

type eventBusImpl struct {
	*data.Data

	helper *log.Helper
}
