package impl

import (
	rabbitwebhookv1 "github.com/aide-family/rabbit/domain/webhook/v1"
	rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewRabbitWebhookRepository(c *conf.Bootstrap, d *data.Data) (repository.RabbitWebhook, error) {
	repoConfig := c.GetWebhookConfig()
	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		factory, ok := rabbitwebhookv1.GetWebhookV1Factory(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("rabbit webhook repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig)
		if err != nil {
			return nil, err
		}
		d.AppendClose("rabbitWebhookRepo", close)
		return &rabbitWebhookRepository{WebhookServer: repoImpl}, nil
	}
}

type rabbitWebhookRepository struct {
	rabbitv1.WebhookServer
}

