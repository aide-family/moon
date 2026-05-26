// Package alert provides domain factory registration for the alert service.
package alert

import (
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"
	"google.golang.org/protobuf/types/known/anypb"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

var globalRegistry = newRegistry()

func newRegistry() *registry {
	return &registry{
		alertV1: domainregister.NewRegistry[AlertFactoryV1](),
	}
}

type AlertFactoryV1 func(c *config.DomainConfig, driver *anypb.Any) (apiv1.AlertServer, func() error, error)

type registry struct {
	alertV1 *domainregister.Registry[AlertFactoryV1]
}

func RegisterAlertV1Factory(name config.DomainConfig_Driver, factory AlertFactoryV1) {
	globalRegistry.alertV1.Register(name, factory)
}

func GetAlertV1Factory(name config.DomainConfig_Driver) (AlertFactoryV1, bool) {
	return globalRegistry.alertV1.Get(name)
}
