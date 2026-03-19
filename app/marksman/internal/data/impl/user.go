package impl

import (
	"github.com/aide-family/magicbox/merr"
	userDomain "github.com/aide-family/goddess/domain/user"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewUserRepository(c *conf.Bootstrap, d *data.Data) (repository.User, error) {
	repoConfig := c.GetUserDomain()
	if repoConfig == nil {
		return nil, merr.ErrorInternalServer("userDomain is required")
	}
	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		factory, ok := userDomain.GetUserV1Factory(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("user repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig, c.GetDomainDriver())
		if err != nil {
			return nil, err
		}
		d.AppendClose("userRepo", close)
		return &userRepository{UserServer: repoImpl}, nil
	}
}

type userRepository struct {
	goddessv1.UserServer
}
