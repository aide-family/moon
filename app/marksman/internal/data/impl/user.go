package impl

import (
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewUserRepository(c *conf.Bootstrap, d *data.Data) (repository.User, error) {
	repoImpl, close, err := newGoddessUser(c.GetGoddessDomain())
	if err != nil {
		return nil, err
	}
	d.AppendClose("userRepo", close)
	return &userRepository{UserServer: repoImpl}, nil
}

type userRepository struct {
	goddessv1.UserServer
}
