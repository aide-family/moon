package impl

import (
	rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewRabbitWebhookRepository(c *conf.Bootstrap, d *data.Data) (repository.RabbitWebhook, error) {
	repoImpl, close, err := newRabbitWebhook(c.GetRabbitDomain())
	if err != nil {
		return nil, err
	}
	d.AppendClose("rabbitWebhookRepo", close)
	return &rabbitWebhookRepository{WebhookServer: repoImpl}, nil
}

type rabbitWebhookRepository struct {
	rabbitv1.WebhookServer
}
