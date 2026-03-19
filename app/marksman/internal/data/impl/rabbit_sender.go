package impl

import (
	"github.com/aide-family/magicbox/merr"
	senderDomain "github.com/aide-family/rabbit/domain/sender"
	rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewRabbitSenderRepository(c *conf.Bootstrap, d *data.Data) (repository.RabbitSender, error) {
	repoConfig := c.GetSenderDomain()
	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		factory, ok := senderDomain.GetSenderV1Factory(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("rabbit sender repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig, c.GetRabbitDomainDriver())
		if err != nil {
			return nil, err
		}
		d.AppendClose("rabbitSenderRepo", close)
		return &rabbitSenderRepository{SenderServer: repoImpl}, nil
	}
}

type rabbitSenderRepository struct {
	rabbitv1.SenderServer
}
