package impl

import (
	namespaceDomain "github.com/aide-family/goddess/domain/namespace"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewNamespaceRepository(c *conf.Bootstrap, d *data.Data) (repository.Namespace, error) {
	repoConfig := c.GetNamespaceDomain()
	if repoConfig == nil {
		return nil, merr.ErrorInternalServer("namespaceDomain is required")
	}
	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		factory, ok := namespaceDomain.GetNamespaceV1Factory(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("namespace repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig, c.GetGoddessDomainDriver())
		if err != nil {
			return nil, err
		}
		d.AppendClose("namespaceRepo", close)
		return &namespaceRepository{NamespaceServer: repoImpl}, nil
	}
}

type namespaceRepository struct {
	goddessv1.NamespaceServer
}
