package biz

import (
	"context"

	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
	"github.com/aide-family/magicbox/merr"
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

func (n *Namespace) CreateNamespace(ctx context.Context, req *bo.CreateNamespaceBo) error {
	if _, err := n.namespaceRepo.GetNamespaceByName(ctx, req.Name); err == nil {
		return merr.ErrorParams("namespace %s already exists", req.Name)
	} else if !merr.IsNotFound(err) {
		n.helper.Errorw("msg", "check namespace exists failed", "error", err, "name", req.Name)
		return merr.ErrorInternalServer("create namespace %s failed", req.Name).WithCause(err)
	}
	if err := n.namespaceRepo.CreateNamespace(ctx, req); err != nil {
		n.helper.Errorw("msg", "create namespace failed", "error", err, "name", req.Name)
		return merr.ErrorInternalServer("create namespace %s failed", req.Name).WithCause(err)
	}
	return nil
}

func (n *Namespace) UpdateNamespace(ctx context.Context, req *bo.UpdateNamespaceBo) error {
	existNamespace, err := n.namespaceRepo.GetNamespaceByName(ctx, req.Name)
	if err != nil && !merr.IsNotFound(err) {
		n.helper.Errorw("msg", "check namespace exists failed", "error", err, "name", req.Name)
		return merr.ErrorInternalServer("update namespace %s failed", req.Name).WithCause(err)
	} else if existNamespace != nil && existNamespace.UID != req.UID {
		return merr.ErrorParams("namespace %s already exists", req.Name)
	}
	if err := n.namespaceRepo.UpdateNamespace(ctx, req); err != nil {
		n.helper.Errorw("msg", "update namespace failed", "error", err, "uid", req.UID)
		return merr.ErrorInternalServer("update namespace %s failed", req.UID).WithCause(err)
	}
	return nil
}

func (n *Namespace) UpdateNamespaceStatus(ctx context.Context, req *bo.UpdateNamespaceStatusBo) error {
	if err := n.namespaceRepo.UpdateNamespaceStatus(ctx, req); err != nil {
		n.helper.Errorw("msg", "update namespace status failed", "error", err, "uid", req.UID)
		return merr.ErrorInternalServer("update namespace status %s failed", req.UID).WithCause(err)
	}
	return nil
}

func (n *Namespace) DeleteNamespace(ctx context.Context, uid snowflake.ID) error {
	if err := n.namespaceRepo.DeleteNamespace(ctx, uid); err != nil {
		n.helper.Errorw("msg", "delete namespace failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("delete namespace %s failed", uid).WithCause(err)
	}
	return nil
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

func (n *Namespace) ListNamespacesByUIDs(ctx context.Context, uids []snowflake.ID) ([]*bo.NamespaceItemBo, error) {
	items, err := n.namespaceRepo.ListNamespacesByUIDs(ctx, uids)
	if err != nil {
		n.helper.Errorw("msg", "list namespaces by uids failed", "error", err, "uids", uids)
		return nil, merr.ErrorInternalServer("list namespaces failed").WithCause(err)
	}
	return items, nil
}

func (n *Namespace) GetNamespaceByName(ctx context.Context, name string) (*bo.NamespaceItemBo, error) {
	namespaceItemBo, err := n.namespaceRepo.GetNamespaceByName(ctx, name)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("namespace %s not found", name)
		}
		n.helper.Errorw("msg", "get namespace failed", "error", err, "name", name)
		return nil, merr.ErrorInternalServer("get namespace %s failed", name).WithCause(err)
	}
	return namespaceItemBo, nil
}

func (n *Namespace) ListNamespace(ctx context.Context, req *bo.ListNamespaceBo) (*bo.PageResponseBo[*bo.NamespaceItemBo], error) {
	pageResponseBo, err := n.namespaceRepo.ListNamespace(ctx, req)
	if err != nil {
		n.helper.Errorw("msg", "list namespace failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list namespace failed").WithCause(err)
	}
	return bo.NewPageResponseBo(pageResponseBo.PageRequestBo, pageResponseBo.GetItems()), nil
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
