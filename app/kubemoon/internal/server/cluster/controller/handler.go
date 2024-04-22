package controller

import (
	"time"
)

type handler struct {
	handlers map[any]HandlerFunc
}

func newHandler() *handler {
	return &handler{
		handlers: make(map[any]HandlerFunc),
	}
}

func (h *handler) addHandler(handle any, operate HandlerFunc) {
	h.handlers[handle] = operate
}

func (h *handler) getHandler(handle any) HandlerFunc {
	return h.handlers[handle]
}

func (h *handler) handler(c *Context) (*time.Duration, error) {
	return c.Next()
}

func EmptyHandler(c *Context) (*time.Duration, error) {
	return nil, nil
}
