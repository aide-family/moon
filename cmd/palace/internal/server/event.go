package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

var _ transport.Server = (*Event)(nil)

func NewEvent(logger log.Logger) *Event {
	return &Event{
		helper: log.NewHelper(logger),
	}
}

type Event struct {
	helper *log.Helper
}

// Start implements transport.Server.
func (e *Event) Start(context.Context) error {
	defer e.helper.Info("[Event] server started")
	return nil
}

// Stop implements transport.Server.
func (e *Event) Stop(context.Context) error {
	defer e.helper.Info("[Event] server stopped")
	return nil
}
