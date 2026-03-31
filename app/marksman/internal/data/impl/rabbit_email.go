package impl

import (
	"github.com/aide-family/magicbox/merr"
	emailDomain "github.com/aide-family/rabbit/domain/email"
	rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewRabbitEmailRepository(c *conf.Bootstrap, d *data.Data) (repository.RabbitEmail, error) {
	repoConfig := c.GetEmailDomain()
	if repoConfig == nil {
		return nil, merr.ErrorInternalServer("emailDomain is required")
	}

	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		factory, ok := emailDomain.GetEmailV1Factory(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("rabbit email repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig, c.GetRabbitDomainDriver())
		if err != nil {
			return nil, err
		}
		d.AppendClose("rabbitEmailRepo", close)
		return &rabbitEmailRepository{EmailServer: repoImpl}, nil
	}
}

type rabbitEmailRepository struct {
	rabbitv1.EmailServer
}
