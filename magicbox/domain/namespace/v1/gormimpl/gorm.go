// Package gormimpl is the implementation of the gorm repository for the namespace service.
package gormimpl

import (
	"context"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	apiv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/contextx"
	namespacev1 "github.com/aide-family/magicbox/domain/namespace/v1"
	"github.com/aide-family/magicbox/domain/namespace/v1/gormimpl/model"
	"github.com/aide-family/magicbox/domain/namespace/v1/gormimpl/query"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/safety"
	"github.com/aide-family/magicbox/strutil"
)

func init() {
	namespacev1.RegisterNamespaceV1Factory(config.DomainConfig_GORM, NewGormRepository)
}

func NewGormRepository(c *config.DomainConfig) (apiv1.NamespaceServer, func() error, error) {
	ormConfig := &config.ORMConfig{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), ormConfig, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal orm config failed: %v", err)
		}
	}
	db, close, err := connect.NewDB(ormConfig)
	if err != nil {
		return nil, nil, err
	}
	query.SetDefault(db)
	return &gormRepository{repoConfig: c, db: db}, close, nil
}

type gormRepository struct {
	apiv1.UnimplementedNamespaceServer
	repoConfig *config.DomainConfig
	db         *gorm.DB
}

// CreateNamespace implements [namespacev1.Repository].
func (g *gormRepository) CreateNamespace(ctx context.Context, req *apiv1.CreateNamespaceRequest) (*apiv1.CreateNamespaceReply, error) {
	namespace := &model.Namespace{
		Name:     req.Name,
		Metadata: safety.NewMap(req.Metadata),
		Status:   enum.GlobalStatus_ENABLED,
		Remark:   req.Remark,
	}
	namespace.WithCreator(contextx.GetUserUID(ctx))
	if err := query.Namespace.WithContext(ctx).Create(namespace); err != nil {
		return nil, merr.ErrorInternalServer("create namespace failed: %v", err)
	}
	return &apiv1.CreateNamespaceReply{
		Uid: namespace.ID.Int64(),
	}, nil
}

// DeleteNamespace implements [namespacev1.Repository].
func (g *gormRepository) DeleteNamespace(ctx context.Context, req *apiv1.DeleteNamespaceRequest) (*apiv1.DeleteNamespaceReply, error) {
	namespaceMutation := query.Namespace
	_, err := namespaceMutation.WithContext(ctx).Where(namespaceMutation.ID.Eq(req.Uid)).Delete()
	if err != nil {
		return nil, merr.ErrorInternalServer("delete namespace failed: %v", err)
	}
	return &apiv1.DeleteNamespaceReply{}, nil
}

// ListNamespace implements [namespacev1.Repository].
func (g *gormRepository) ListNamespace(ctx context.Context, req *apiv1.ListNamespaceRequest) (*apiv1.ListNamespaceReply, error) {
	namespaceMutation := query.Namespace
	wrappers := namespaceMutation.WithContext(ctx)
	if strutil.IsNotEmpty(req.Keyword) {
		wrappers = wrappers.Where(namespaceMutation.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status > enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(namespaceMutation.Status.Eq(int32(req.Status)))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, merr.ErrorInternalServer("count namespace failed: %v", err)
	}
	if req.Page > 0 && req.PageSize > 0 {
		wrappers = wrappers.Limit(int(req.PageSize)).Offset(int(req.PageSize * (req.Page - 1)))
	}
	wrappers = wrappers.Order(namespaceMutation.ID.Desc())
	queryNamespaces, err := wrappers.Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("list namespace failed: %v", err)
	}
	namespaces := make([]*apiv1.NamespaceItem, 0, len(queryNamespaces))
	for _, queryNamespace := range queryNamespaces {
		namespaces = append(namespaces, ConvertNamespaceItem(queryNamespace))
	}
	return &apiv1.ListNamespaceReply{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Items:    namespaces,
	}, nil
}

// UpdateNamespace implements [namespacev1.Repository].
func (g *gormRepository) UpdateNamespace(ctx context.Context, req *apiv1.UpdateNamespaceRequest) (*apiv1.UpdateNamespaceReply, error) {
	namespaceMutation := query.Namespace
	columns := []field.AssignExpr{
		namespaceMutation.Name.Value(req.Name),
		namespaceMutation.Metadata.Value(safety.NewMap(req.Metadata)),
		namespaceMutation.Remark.Value(req.Remark),
	}
	_, err := namespaceMutation.WithContext(ctx).Where(namespaceMutation.ID.Eq(req.Uid)).UpdateColumnSimple(columns...)
	if err != nil {
		return nil, merr.ErrorInternalServer("update namespace failed: %v", err)
	}
	return &apiv1.UpdateNamespaceReply{}, nil
}

// UpdateNamespaceStatus implements [namespacev1.Repository].
func (g *gormRepository) UpdateNamespaceStatus(ctx context.Context, req *apiv1.UpdateNamespaceStatusRequest) (*apiv1.UpdateNamespaceStatusReply, error) {
	namespaceMutation := query.Namespace
	_, err := namespaceMutation.WithContext(ctx).Where(namespaceMutation.ID.Eq(req.Uid)).UpdateColumnSimple(namespaceMutation.Status.Value(int32(req.Status)))
	if err != nil {
		return nil, merr.ErrorInternalServer("update namespace status failed: %v", err)
	}
	return &apiv1.UpdateNamespaceStatusReply{}, nil
}

func (g *gormRepository) GetNamespace(ctx context.Context, req *apiv1.GetNamespaceRequest) (*apiv1.NamespaceItem, error) {
	namespace, err := query.Namespace.WithContext(ctx).Where(query.Namespace.ID.Eq(req.Uid)).First()
	if err != nil {
		return nil, merr.ErrorInternalServer("get namespace failed: %v", err)
	}
	return ConvertNamespaceItem(namespace), nil
}

// SelectNamespace implements [namespacev1.Repository].
func (g *gormRepository) SelectNamespace(ctx context.Context, req *apiv1.SelectNamespaceRequest) (*apiv1.SelectNamespaceReply, error) {
	mutation := query.Namespace
	wrappers := mutation.WithContext(ctx)
	if strutil.IsNotEmpty(req.Keyword) {
		wrappers = wrappers.Where(mutation.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status > enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(mutation.Status.Eq(int32(req.Status)))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, merr.ErrorInternalServer("count namespace failed: %v", err)
	}
	if req.LastUID > 0 {
		wrappers = wrappers.Where(mutation.ID.Lt(req.LastUID))
	}
	wrappers = wrappers.Limit(int(req.Limit))
	wrappers = wrappers.Select(mutation.ID, mutation.Name, mutation.Status, mutation.DeletedAt, mutation.Remark)
	wrappers = wrappers.Order(mutation.ID.Desc())
	queryNamespaces, err := wrappers.Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("select namespace failed: %v", err)
	}
	namespaces := make([]*apiv1.NamespaceItemSelect, 0, len(queryNamespaces))
	lastUID := int64(0)
	for _, queryNamespace := range queryNamespaces {
		namespace := ConvertNamespaceItemSelect(queryNamespace)
		namespaces = append(namespaces, namespace)
		lastUID = namespace.Value
	}
	return &apiv1.SelectNamespaceReply{
		Items:   namespaces,
		Total:   total,
		LastUID: lastUID,
		HasMore: len(queryNamespaces) == int(req.Limit),
	}, nil
}
