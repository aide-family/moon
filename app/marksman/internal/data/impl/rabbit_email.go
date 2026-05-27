package impl

import (
	rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewRabbitEmailRepository(c *conf.Bootstrap, d *data.Data) (repository.RabbitEmail, error) {
	repoImpl, close, err := newRabbitEmail(c.GetRabbitDomain())
	if err != nil {
		return nil, err
	}
	d.AppendClose("rabbitEmailRepo", close)
	return &rabbitEmailRepository{EmailServer: repoImpl}, nil
}

type rabbitEmailRepository struct {
	rabbitv1.EmailServer
}
