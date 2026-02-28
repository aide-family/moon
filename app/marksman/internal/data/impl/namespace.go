package impl

import (
	"context"

	apiv1 "github.com/aide-family/magicbox/api/v1"
	namespacev1 "github.com/aide-family/magicbox/domain/namespace/v1"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
	"github.com/aide-family/marksman/internal/data/impl/convert"
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
		return &namespaceRepository{repo: repoImpl}, nil
	}
}

type namespaceRepository struct {
	repo apiv1.NamespaceServer
}

// GetNamespace implements [repository.Namespace].
func (n *namespaceRepository) GetNamespace(ctx context.Context, uid snowflake.ID) (*bo.NamespaceItemBo, error) {
	namespaceModel, err := n.repo.GetNamespace(ctx, &apiv1.GetNamespaceRequest{
		Uid: uid.Int64(),
	})
	if err != nil {
		return nil, err
	}
	return convert.ToNamespaceItemBo(namespaceModel), nil
}

// SelectNamespace implements [repository.Namespace].
func (n *namespaceRepository) SelectNamespace(ctx context.Context, req *bo.SelectNamespaceBo) (*bo.SelectNamespaceBoResult, error) {
	selectNamespaceResponse, err := n.repo.SelectNamespace(ctx, &apiv1.SelectNamespaceRequest{
		Keyword: req.Keyword,
		Limit:   req.Limit,
		LastUID: req.LastUID.Int64(),
		Status:  enum.GlobalStatus(req.Status),
	})
	if err != nil {
		return nil, err
	}
	items := make([]*bo.NamespaceItemSelectBo, 0, len(selectNamespaceResponse.Items))
	for _, namespaceItemSelect := range selectNamespaceResponse.Items {
		items = append(items, convert.ToNamespaceItemSelectBo(namespaceItemSelect))
	}
	return &bo.SelectNamespaceBoResult{
		Items:   items,
		Total:   selectNamespaceResponse.Total,
		LastUID: snowflake.ParseInt64(selectNamespaceResponse.LastUID),
		HasMore: selectNamespaceResponse.HasMore,
	}, nil
}
