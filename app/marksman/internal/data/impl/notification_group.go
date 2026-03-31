package impl

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data"
	"github.com/aide-family/marksman/internal/data/impl/convert"
	"github.com/aide-family/marksman/internal/data/impl/query"
)

func NewNotificationGroupRepository(d *data.Data) (repository.NotificationGroup, error) {
	query.SetDefault(d.DB())
	return &notificationGroupRepository{db: d.DB()}, nil
}

type notificationGroupRepository struct {
	db *gorm.DB
}

func (r *notificationGroupRepository) CreateNotificationGroup(ctx context.Context, req *bo.CreateNotificationGroupBo) (snowflake.ID, error) {
	m := convert.ToNotificationGroupDo(ctx, req)
	if err := query.NotificationGroup.WithContext(ctx).Create(m); err != nil {
		return 0, err
	}
	return m.ID, nil
}

func (r *notificationGroupRepository) NotificationGroupNameTaken(ctx context.Context, name string, excludeUID snowflake.ID) (bool, error) {
	n := query.NotificationGroup
	total, err := n.WithContext(ctx).Where(
		n.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		n.Name.Eq(name),
		n.ID.Neq(excludeUID.Int64()),
	).Count()
	if err != nil {
		return false, err
	}
	return total > 0, nil
}

func (r *notificationGroupRepository) UpdateNotificationGroup(ctx context.Context, req *bo.UpdateNotificationGroupBo) error {
	n := query.NotificationGroup
	doUpdate := convert.ToNotificationGroupDoUpdate(req)
	columns := []field.AssignExpr{
		n.Name.Value(req.Name),
		n.Remark.Value(req.Remark),
		n.Metadata.Value(doUpdate.Metadata),
		n.Members.Value(doUpdate.Members),
		n.Webhooks.Value(doUpdate.Webhooks),
		n.Templates.Value(doUpdate.Templates),
		n.EmailConfigs.Value(doUpdate.EmailConfigs),
	}
	_, err := n.WithContext(ctx).Where(
		n.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		n.ID.Eq(req.UID.Int64()),
	).UpdateColumnSimple(columns...)
	if err != nil {
		return err
	}
	return nil
}

func (r *notificationGroupRepository) UpdateNotificationGroupStatus(ctx context.Context, req *bo.UpdateNotificationGroupStatusBo) error {
	n := query.NotificationGroup
	_, err := n.WithContext(ctx).Where(
		n.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		n.ID.Eq(req.UID.Int64()),
	).Update(n.Status, req.Status)
	if err != nil {
		return err
	}
	return nil
}

func (r *notificationGroupRepository) DeleteNotificationGroup(ctx context.Context, uid snowflake.ID) error {
	n := query.NotificationGroup
	_, err := n.WithContext(ctx).Where(
		n.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		n.ID.Eq(uid.Int64()),
	).Delete()
	if err != nil {
		return err
	}
	return nil
}

func (r *notificationGroupRepository) GetNotificationGroup(ctx context.Context, uid snowflake.ID) (*bo.NotificationGroupItemBo, error) {
	n := query.NotificationGroup
	m, err := n.WithContext(ctx).Where(
		n.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		n.ID.Eq(uid.Int64()),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("notification group not found")
		}
		return nil, err
	}
	return convert.ToNotificationGroupItemBo(m), nil
}

func (r *notificationGroupRepository) ListNotificationGroup(ctx context.Context, req *bo.ListNotificationGroupBo) (*bo.PageResponseBo[*bo.NotificationGroupItemBo], error) {
	n := query.NotificationGroup
	wrappers := n.WithContext(ctx).Where(n.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()))
	if req.Keyword != "" {
		wrappers = wrappers.Where(n.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status != enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(n.Status.Eq(int32(req.Status)))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, err
	}
	req.WithTotal(total)
	if req.Page > 0 && req.PageSize > 0 {
		wrappers = wrappers.Offset(req.Offset()).Limit(req.Limit())
	}
	list, err := wrappers.Order(n.UpdatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}
	items := make([]*bo.NotificationGroupItemBo, 0, len(list))
	for _, m := range list {
		items = append(items, convert.ToNotificationGroupItemBo(m))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}
