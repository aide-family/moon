package impl

import (
	magicboxapiv1 "github.com/aide-family/magicbox/api/v1"
	namespacev1 "github.com/aide-family/magicbox/domain/namespace/v1"
	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
)

func NewNamespaceRepository(c *conf.Bootstrap, d *data.Data) (repository.Namespace, error) {
	repoConfig := c.GetNamespaceConfig()
	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		factory, ok := namespacev1.GetNamespaceV1Factory(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("namespace repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig)
		if err != nil {
			return nil, err
		}
		d.AppendClose("namespaceRepo", close)
		return &namespaceRepository{NamespaceServer: repoImpl}, nil
	}
}

type namespaceRepository struct {
	magicboxapiv1.NamespaceServer
}
