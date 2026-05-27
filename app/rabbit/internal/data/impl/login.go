package impl

import (
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"

	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
)

type loginRepository struct {
	goddessv1.AuthServiceServer
}

func NewLoginRepository(c *conf.Bootstrap, d *data.Data) (repository.LoginRepository, error) {
	repoImpl, close, err := newGoddessAuth(c.GetGoddessDomain())
	if err != nil {
		return nil, err
	}
	d.AppendClose("loginRepo", close)
	return &loginRepository{AuthServiceServer: repoImpl}, nil
}
