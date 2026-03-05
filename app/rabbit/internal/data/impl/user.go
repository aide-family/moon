package impl

import (
	userv1 "github.com/aide-family/goddess/domain/user/v1"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
)

func NewUserRepository(c *conf.Bootstrap, d *data.Data) (repository.User, error) {
	repoConfig := c.GetUserConfig()
	if repoConfig == nil {
		return nil, merr.ErrorInternalServer("userConfig is required")
	}
	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		factory, ok := userv1.GetUserFactoryV1(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("user repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig)
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
