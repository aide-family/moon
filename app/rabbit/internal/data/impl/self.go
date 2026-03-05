package impl

import (
	selfv1 "github.com/aide-family/goddess/domain/self/v1"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
)

func NewSelfRepository(c *conf.Bootstrap, d *data.Data) (repository.Self, error) {
	repoConfig := c.GetSelfConfig()
	if repoConfig == nil {
		return nil, merr.ErrorInternalServer("selfConfig is required")
	}
	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		factory, ok := selfv1.GetSelfFactoryV1(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("self repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig)
		if err != nil {
			return nil, err
		}
		d.AppendClose("selfRepo", close)
		return &selfRepository{SelfServer: repoImpl}, nil
	}
}

type selfRepository struct {
	goddessv1.SelfServer
}
