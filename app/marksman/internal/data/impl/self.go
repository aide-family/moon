package impl

import (
	selfDomain "github.com/aide-family/goddess/domain/self"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewSelfRepository(c *conf.Bootstrap, d *data.Data) (repository.Self, error) {
	repoConfig := c.GetSelfDomain()
	if repoConfig == nil {
		return nil, merr.ErrorInternalServer("selfDomain is required")
	}
	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		factory, ok := selfDomain.GetSelfV1Factory(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("self repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig, c.GetGoddessDomainDriver())
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
