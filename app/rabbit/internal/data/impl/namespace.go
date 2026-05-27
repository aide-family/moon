package impl

import (
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"

	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
)

func NewNamespaceRepository(c *conf.Bootstrap, d *data.Data) (repository.Namespace, error) {
	repoImpl, close, err := newGoddessNamespace(c.GetGoddessDomain())
	if err != nil {
		return nil, err
	}
	d.AppendClose("namespaceRepo", close)
	return &namespaceRepository{NamespaceServer: repoImpl}, nil
}

type namespaceRepository struct {
	goddessv1.NamespaceServer
}
