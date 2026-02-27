package impl

import (
	"context"
	"errors"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/safety"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
	"github.com/aide-family/goddess/internal/data"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	"github.com/aide-family/goddess/internal/data/impl/query"
)

func NewNamespaceRepository(d *data.Data) repository.Namespace {
	return &namespaceRepository{Data: d}
}

type namespaceRepository struct {
	*data.Data
}

// AllNamespaces implements [repository.Namespace].
func (n *namespaceRepository) AllNamespaces(ctx context.Context) ([]*bo.NamespaceItemBo, error) {
	namespaceMutation := query.Namespace
	namespaces, err := namespaceMutation.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}
	items := make([]*bo.NamespaceItemBo, 0, len(namespaces))
	for _, namespace := range namespaces {
		items = append(items, convert.NamespaceToBo(namespace))
	}
	return items, nil
}

// CreateNamespace implements [repository.Namespace].
func (n *namespaceRepository) CreateNamespace(ctx context.Context, req *bo.CreateNamespaceBo) error {
	namespaceModel := convert.NamespaceToDo(ctx, req)
	mutation := query.Namespace
	return mutation.WithContext(ctx).Create(namespaceModel)
}

// DeleteNamespace implements [repository.Namespace].
func (n *namespaceRepository) DeleteNamespace(ctx context.Context, uid snowflake.ID) error {
	mutation := query.Namespace
	_, err := mutation.WithContext(ctx).Where(mutation.UID.Eq(uid.Int64())).Delete()
	return err
}

// GetNamespace implements [repository.Namespace].
func (n *namespaceRepository) GetNamespace(ctx context.Context, uid snowflake.ID) (*bo.NamespaceItemBo, error) {
	mutation := query.Namespace
	namespace, err := mutation.WithContext(ctx).Where(mutation.UID.Eq(uid.Int64())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("namespace %s not found", uid)
		}
		return nil, err
	}
	return convert.NamespaceToBo(namespace), nil
}

// ListNamespacesByUIDs implements [repository.Namespace].
func (n *namespaceRepository) ListNamespacesByUIDs(ctx context.Context, uids []snowflake.ID) ([]*bo.NamespaceItemBo, error) {
	if len(uids) == 0 {
		return nil, nil
	}
	ids := make([]int64, 0, len(uids))
	for _, uid := range uids {
		ids = append(ids, uid.Int64())
	}
	mutation := query.Namespace
	namespaces, err := mutation.WithContext(ctx).Where(mutation.UID.In(ids...)).Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("list namespaces by uids failed: %v", err)
	}
	items := make([]*bo.NamespaceItemBo, 0, len(namespaces))
	for _, ns := range namespaces {
		items = append(items, convert.NamespaceToBo(ns))
	}
	return items, nil
}

// GetNamespaceByName implements [repository.Namespace].
func (n *namespaceRepository) GetNamespaceByName(ctx context.Context, name string) (*bo.NamespaceItemBo, error) {
	mutation := query.Namespace
	namespace, err := mutation.WithContext(ctx).Where(mutation.Name.Eq(name)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("namespace %s not found", name)
		}
		return nil, err
	}
	return convert.NamespaceToBo(namespace), nil
}

// ListNamespace implements [repository.Namespace].
func (n *namespaceRepository) ListNamespace(ctx context.Context, req *bo.ListNamespaceBo) (*bo.PageResponseBo[*bo.NamespaceItemBo], error) {
	mutation := query.Namespace
	wrappers := mutation.WithContext(ctx)
	if strutil.IsNotEmpty(req.Keyword) {
		wrappers = wrappers.Where(mutation.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status > enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(mutation.Status.Eq(int32(req.Status)))
	}

	if req.Page > 0 && req.PageSize > 0 {
		total, err := wrappers.Count()
		if err != nil {
			return nil, merr.ErrorInternalServer("list namespace failed: %v", err)
		}
		req.WithTotal(total)
		wrappers = wrappers.Limit(int(req.PageSize)).Offset(int((req.Page - 1) * req.PageSize))
	}
	queryNamespaces, err := wrappers.Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("list namespace failed: %v", err)
	}
	items := make([]*bo.NamespaceItemBo, 0, len(queryNamespaces))
	for _, namespace := range queryNamespaces {
		items = append(items, convert.NamespaceToBo(namespace))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

// SelectNamespace implements [repository.Namespace].
func (n *namespaceRepository) SelectNamespace(ctx context.Context, req *bo.SelectNamespaceBo) (*bo.SelectNamespaceBoResult, error) {
	mutation := query.Namespace
	wrappers := mutation.WithContext(ctx)
	if strutil.IsNotEmpty(req.Keyword) {
		wrappers = wrappers.Where(mutation.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status > enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(mutation.Status.Eq(int32(req.Status)))
	}

	if req.LastUID > 0 {
		wrappers = wrappers.Where(mutation.UID.Lt(req.LastUID.Int64()))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, merr.ErrorInternalServer("select namespace failed: %v", err)
	}
	wrappers = wrappers.Limit(int(req.Limit))
	wrappers = wrappers.Select(mutation.UID, mutation.Name, mutation.Status, mutation.DeletedAt)
	queryNamespaces, err := wrappers.Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("select namespace failed: %v", err)
	}
	items := make([]*bo.NamespaceItemSelectBo, 0, len(queryNamespaces))
	for _, namespace := range queryNamespaces {
		items = append(items, convert.NamespaceToSelectBo(namespace))
	}
	return &bo.SelectNamespaceBoResult{
		Items:   items,
		Total:   total,
		LastUID: req.LastUID,
		HasMore: len(items) >= int(req.Limit),
	}, nil
}

// UpdateNamespace implements [repository.Namespace].
func (n *namespaceRepository) UpdateNamespace(ctx context.Context, req *bo.UpdateNamespaceBo) error {
	mutation := query.Namespace
	wrappers := []gen.Condition{
		mutation.UID.Eq(req.UID.Int64()),
	}
	columns := []field.AssignExpr{
		mutation.Name.Value(req.Name),
		mutation.Metadata.Value(safety.NewMap(req.Metadata)),
	}

	_, err := mutation.WithContext(ctx).Where(wrappers...).UpdateColumnSimple(columns...)
	return err
}

// UpdateNamespaceStatus implements [repository.Namespace].
func (n *namespaceRepository) UpdateNamespaceStatus(ctx context.Context, req *bo.UpdateNamespaceStatusBo) error {
	mutation := query.Namespace
	_, err := mutation.WithContext(ctx).Where(mutation.UID.Eq(req.UID.Int64())).Update(mutation.Status, req.Status)
	return err
}
