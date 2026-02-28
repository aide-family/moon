package biz

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewNamespace(
	namespaceRepo repository.Namespace,
	helper *klog.Helper,
) *Namespace {
	return &Namespace{
		namespaceRepo: namespaceRepo,
		helper:        klog.NewHelper(klog.With(helper.Logger(), "biz", "namespace")),
	}
}

type Namespace struct {
	helper        *klog.Helper
	namespaceRepo repository.Namespace
}

func (n *Namespace) GetNamespace(ctx context.Context, uid snowflake.ID) (*bo.NamespaceItemBo, error) {
	namespaceItemBo, err := n.namespaceRepo.GetNamespace(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("namespace %s not found", uid)
		}

		n.helper.Errorw("msg", "get namespace failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get namespace %s failed", uid).WithCause(err)
	}
	return namespaceItemBo, nil
}

func (n *Namespace) SelectNamespace(ctx context.Context, req *bo.SelectNamespaceBo) (*bo.SelectNamespaceBoResult, error) {
	result, err := n.namespaceRepo.SelectNamespace(ctx, req)
	if err != nil {
		n.helper.Errorw("msg", "select namespace failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("select namespace failed").WithCause(err)
	}
	return &bo.SelectNamespaceBoResult{
		Items:   result.Items,
		Total:   result.Total,
		LastUID: result.LastUID,
	}, nil
}
