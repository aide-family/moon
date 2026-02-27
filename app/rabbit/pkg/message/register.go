package message

import (
	"context"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
)

var globalRegistry = NewRegistry()

func NewRegistry() Registry {
	return &registry{
		drivers: safety.NewSyncMap(make(map[enum.MessageType]Driver)),
	}
}

type (
	Sender interface {
		Send(ctx context.Context, message Message) error
	}

	Driver func(c *config.MessageConfig) (Sender, error)
)

type Registry interface {
	RegisterDriver(messageType enum.MessageType, driver Driver)
	GetDriver(messageType enum.MessageType) (Driver, bool)
}

type registry struct {
	drivers *safety.SyncMap[enum.MessageType, Driver]
}

// GetDriver implements [Registry].
func (r *registry) GetDriver(messageType enum.MessageType) (Driver, bool) {
	return r.drivers.Get(messageType)
}

// RegisterDriver implements [Registry].
func (r *registry) RegisterDriver(messageType enum.MessageType, driver Driver) {
	r.drivers.Set(messageType, driver)
}

func RegisterDriver(messageType enum.MessageType, driver Driver) {
	globalRegistry.RegisterDriver(messageType, driver)
}

func GetDriver(messageType enum.MessageType) (Driver, bool) {
	return globalRegistry.GetDriver(messageType)
}
