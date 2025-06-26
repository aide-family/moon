package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/moon/cmd/palace/internal/service"
	"github.com/aide-family/moon/pkg/util/safety"
)

var _ transport.Server = (*Event)(nil)

func NewEvent(
	loadService *service.LoadService,
	serverService *service.ServerService,
	logger log.Logger,
) *Event {
	return &Event{
		helper:        log.NewHelper(logger),
		loadService:   loadService,
		serverService: serverService,
	}
}

type Event struct {
	helper        *log.Helper
	loadService   *service.LoadService
	serverService *service.ServerService
}

// Start implements transport.Server.
func (e *Event) Start(context.Context) error {
	defer e.helper.Info("[Event] server started")
	safety.Go("subscribeDataChangeEvent", func() {
		for event := range e.loadService.SubscribeDataChangeEvent() {
			e.helper.Debugf("[Event] received data change event: %v", event)
			if err := e.serverService.Sync(context.Background(), event); err != nil {
				e.helper.Errorf("[Event] sync data change event error: %v", event, err)
			}
		}
	}, e.helper.Logger())
	return nil
}

// Stop implements transport.Server.
func (e *Event) Stop(context.Context) error {
	defer e.helper.Info("[Event] server stopped")
	return nil
}
