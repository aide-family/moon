package impl

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/merr"
	alertDomain "github.com/aide-family/rabbit/domain/alert"
	alertv1 "github.com/aide-family/rabbit/domain/alert/v1"
	rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewRabbitAlertRepository(c *conf.Bootstrap, d *data.Data) (repository.RabbitAlert, error) {
	repoConfig := c.GetAlertDomain()
	if repoConfig == nil {
		return nil, merr.ErrorInternalServer("alertDomain is required")
	}

	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		if driver == config.DomainConfig_DEFAULT {
			repoImpl, close, err := alertv1.NewDefaultAlert(repoConfig, c.GetRabbitDomainDriver(), c.GetGoddessDomainDriver(), c.GetMemberDomain())
			if err != nil {
				return nil, err
			}
			d.AppendClose("rabbitAlertRepo", close)
			return &rabbitAlertRepository{AlertServer: repoImpl}, nil
		}
		factory, ok := alertDomain.GetAlertV1Factory(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("rabbit alert repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig, c.GetRabbitDomainDriver())
		if err != nil {
			return nil, err
		}
		d.AppendClose("rabbitAlertRepo", close)
		return &rabbitAlertRepository{AlertServer: repoImpl}, nil
	}
}

type rabbitAlertRepository struct {
	rabbitv1.AlertServer
}
