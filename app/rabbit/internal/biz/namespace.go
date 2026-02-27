package biz

import (
	"context"

	magicboxapiv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/rabbit/internal/biz/repository"
)

func NewNamespace(
	namespaceRepo repository.Namespace,
	helper *klog.Helper,
) *Namespace {
	return &Namespace{
		Namespace: namespaceRepo,
		helper:    klog.NewHelper(klog.With(helper.Logger(), "biz", "namespace")),
	}
}

type Namespace struct {
	helper *klog.Helper
	repository.Namespace
}

func (n *Namespace) HasNamespace(ctx context.Context) (snowflake.ID, error) {
	namespace := contextx.GetNamespace(ctx)
	if namespace <= 0 {
		return 0, merr.ErrorForbidden("namespace is required, please set the namespace in the request header or metadata, Example: %s: default", cnst.HTTPHeaderXNamespace)
	}
	req := &magicboxapiv1.GetNamespaceRequest{
		Uid: namespace.Int64(),
	}
	namespaceItemBo, err := n.GetNamespace(ctx, req)
	if err != nil {
		if merr.IsNotFound(err) {
			return 0, merr.ErrorForbidden("namespace %s not allowed", namespace)
		}
		return 0, err
	}
	if namespaceItemBo.Status != enum.GlobalStatus_ENABLED {
		return 0, merr.ErrorForbidden("namespace %s is not allowed", namespace)
	}
	return snowflake.ParseInt64(namespaceItemBo.Uid), nil
}
