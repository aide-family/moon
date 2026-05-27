package impl

import (
	rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewRabbitTemplateRepository(c *conf.Bootstrap, d *data.Data) (repository.RabbitTemplate, error) {
	repoImpl, close, err := newRabbitTemplate(c.GetRabbitDomain())
	if err != nil {
		return nil, err
	}
	d.AppendClose("rabbitTemplateRepo", close)
	return &rabbitTemplateRepository{TemplateServer: repoImpl}, nil
}

type rabbitTemplateRepository struct {
	rabbitv1.TemplateServer
}
