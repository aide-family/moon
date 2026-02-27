package connect

import (
	"gorm.io/gorm"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/safety"
)

var globalRegistry = NewRegistry()

type (
	ORMFactory    func(c *config.ORMConfig) (*gorm.DB, error)
	ReportFactory func(c *config.ReportConfig) (Report, func() error, error)
)

type Registry struct {
	ormConfigs    *safety.SyncMap[config.ORMConfig_Dialector, ORMFactory]
	reportConfigs *safety.SyncMap[config.ReportConfig_ReportType, ReportFactory]
}

func (r *Registry) RegisterORMFactory(dialector config.ORMConfig_Dialector, factory ORMFactory) {
	r.ormConfigs.Set(dialector, factory)
}

func (r *Registry) GetORMFactory(dialector config.ORMConfig_Dialector) (ORMFactory, bool) {
	return r.ormConfigs.Get(dialector)
}

func (r *Registry) RegisterReportFactory(registryType config.ReportConfig_ReportType, factory ReportFactory) {
	r.reportConfigs.Set(registryType, factory)
}

func (r *Registry) GetReportFactory(registryType config.ReportConfig_ReportType) (ReportFactory, bool) {
	return r.reportConfigs.Get(registryType)
}

func NewRegistry() *Registry {
	return &Registry{
		ormConfigs:    safety.NewSyncMap(make(map[config.ORMConfig_Dialector]ORMFactory)),
		reportConfigs: safety.NewSyncMap(make(map[config.ReportConfig_ReportType]ReportFactory)),
	}
}
