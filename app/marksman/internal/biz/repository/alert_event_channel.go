package repository

import "github.com/aide-family/marksman/internal/biz/bo"

// AlertEventChannel is a channel that collects alert events from metric evaluators.
type AlertEventChannel interface {
	// Send sends an alert event (non-blocking best-effort; drops if full).
	Send(event *bo.AlertEventBo)
	// GetChannel returns the read-only channel for consumers.
	GetChannel() <-chan *bo.AlertEventBo
}
