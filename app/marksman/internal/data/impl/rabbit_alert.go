package impl

import (
	rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewRabbitAlertRepository(c *conf.Bootstrap, d *data.Data) (repository.RabbitAlert, error) {
	repoImpl, close, err := newRabbitAlert(c.GetRabbitDomain())
	if err != nil {
		return nil, err
	}
	d.AppendClose("rabbitAlertRepo", close)
	return &rabbitAlertRepository{AlertServer: repoImpl}, nil
}

type rabbitAlertRepository struct {
	rabbitv1.AlertServer
}
