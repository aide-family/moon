package data

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/pointer"
	"github.com/google/wire"

	"github.com/aide-family/jade_tree/internal/conf"
)

var ProviderSetData = wire.NewSet(New)

func New(c *conf.Bootstrap) (*Data, func(), error) {
	d := &Data{}
	if report := c.GetReport(); !pointer.IsNil(report) && report.GetReportType() != config.ReportConfig_REPORT_TYPE_UNKNOWN {
		registry, closer, err := connect.NewReport(report)
		if err != nil {
			return nil, func() {}, err
		}
		d.registry = registry
		return d, func() { _ = closer() }, nil
	}
	return d, func() {}, nil
}

type Data struct {
	registry connect.Report
}

func (d *Data) Registry() connect.Report { return d.registry }
