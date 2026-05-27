package impl

import (
	rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewRabbitSenderRepository(c *conf.Bootstrap, d *data.Data) (repository.RabbitSender, error) {
	repoImpl, close, err := newRabbitSender(c.GetRabbitDomain())
	if err != nil {
		return nil, err
	}
	d.AppendClose("rabbitSenderRepo", close)
	return &rabbitSenderRepository{SenderServer: repoImpl}, nil
}

type rabbitSenderRepository struct {
	rabbitv1.SenderServer
}
