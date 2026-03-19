package impl

import (
	templateDomain "github.com/aide-family/rabbit/domain/template"
	rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewRabbitTemplateRepository(c *conf.Bootstrap, d *data.Data) (repository.RabbitTemplate, error) {
	repoConfig := c.GetMessageTemplateDomain()
	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		factory, ok := templateDomain.GetTemplateV1Factory(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("rabbit template repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig, c.GetRabbitDomainDriver())
		if err != nil {
			return nil, err
		}
		d.AppendClose("rabbitTemplateRepo", close)
		return &rabbitTemplateRepository{TemplateServer: repoImpl}, nil
	}
}

type rabbitTemplateRepository struct {
	rabbitv1.TemplateServer
}
