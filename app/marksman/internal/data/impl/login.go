package impl

import (
	"github.com/aide-family/magicbox/merr"
	authDomain "github.com/aide-family/goddess/domain/auth"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

type loginRepository struct {
	goddessv1.AuthServiceServer
}

func NewLoginRepository(c *conf.Bootstrap, d *data.Data) (repository.LoginRepository, error) {
	repoConfig := c.GetAuthDomain()
	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		factory, ok := authDomain.GetAuthV1Factory(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("auth repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig, c.GetDomainDriver())
		if err != nil {
			return nil, err
		}
		d.AppendClose("loginRepo", close)

		return &loginRepository{AuthServiceServer: repoImpl}, nil
	}
}
