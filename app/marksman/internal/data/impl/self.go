package impl

import (
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewSelfRepository(c *conf.Bootstrap, d *data.Data) (repository.Self, error) {
	repoImpl, close, err := newGoddessSelf(c.GetGoddessDomain())
	if err != nil {
		return nil, err
	}
	d.AppendClose("selfRepo", close)
	return &selfRepository{SelfServer: repoImpl}, nil
}

type selfRepository struct {
	goddessv1.SelfServer
}
