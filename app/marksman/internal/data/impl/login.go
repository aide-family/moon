package impl

import (
	authv1 "github.com/aide-family/goddess/domain/auth/v1"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

type loginRepository struct {
	goddessv1.AuthServiceServer
}

func NewLoginRepository(c *conf.Bootstrap, d *data.Data) (repository.LoginRepository, error) {
	repoConfig := c.GetLoginConfig()
	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		factory, ok := authv1.GetAuthV1Factory(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("auth repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig)
		if err != nil {
			return nil, err
		}
		d.AppendClose("loginRepo", close)

		return &loginRepository{AuthServiceServer: repoImpl}, nil
	}
}
