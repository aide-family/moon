package data

import (
	"context"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertpage"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/useralertpage"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/pkg/merr"
)

func (r *alertPageRepo) CreateAlertPage(ctx context.Context, in bo.AlertPage) (*bo.AlertPage, error) {
	row, err := r.client.AlertPage.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetOwnerUID(in.OwnerUID).
		SetLabels(in.Labels).
		SetPublic(in.Public).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toAlertPage(row)
	return &item, nil
}

func (r *alertPageRepo) UpdateAlertPage(ctx context.Context, in bo.AlertPage) (*bo.AlertPage, error) {
	n, err := r.client.AlertPage.Update().Where(
		alertpage.NamespaceUIDEQ(in.NamespaceUID),
		alertpage.UIDEQ(in.UID),
		alertpage.DeletedAtIsNil(),
	).
		SetName(in.Name).
		SetLabels(in.Labels).
		SetPublic(in.Public).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetAlertPage(ctx, in.NamespaceUID, in.UID)
}

func (r *alertPageRepo) DeleteAlertPage(ctx context.Context, namespaceUID, uid string) error {
	n, err := r.client.AlertPage.Update().Where(
		alertpage.NamespaceUIDEQ(namespaceUID),
		alertpage.UIDEQ(uid),
		alertpage.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	_, err = r.client.UserAlertPage.Delete().Where(
		useralertpage.NamespaceUIDEQ(namespaceUID),
		useralertpage.AlertPageUIDEQ(uid),
	).Exec(ctx)
	if err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *alertPageRepo) GetAlertPage(ctx context.Context, namespaceUID, uid string) (*bo.AlertPage, error) {
	row, err := r.client.AlertPage.Query().Where(
		alertpage.NamespaceUIDEQ(namespaceUID),
		alertpage.UIDEQ(uid),
		alertpage.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toAlertPage(row)
	return &item, nil
}

func (r *alertPageRepo) ListAlertPages(ctx context.Context, namespaceUID, ownerUID, scope, name string, page, pageSize int) ([]bo.AlertPage, int, error) {
	query := r.client.AlertPage.Query().Where(alertpage.NamespaceUIDEQ(namespaceUID), alertpage.DeletedAtIsNil())
	switch scope {
	case "selectable":
		query = query.Where(alertpage.Or(
			alertpage.OwnerUIDEQ(ownerUID),
			alertpage.PublicEQ(true),
		))
	default:
		query = query.Where(alertpage.OwnerUIDEQ(ownerUID))
	}
	if name != "" {
		query = query.Where(alertpage.NameContainsFold(name))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(alertpage.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.AlertPage, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAlertPage(row))
	}
	return items, total, nil
}

func (r *alertPageRepo) ListSelectableByUIDs(ctx context.Context, namespaceUID, ownerUID string, uids []string) ([]bo.AlertPage, error) {
	if len(uids) == 0 {
		return nil, nil
	}
	rows, err := r.client.AlertPage.Query().Where(
		alertpage.NamespaceUIDEQ(namespaceUID),
		alertpage.DeletedAtIsNil(),
		alertpage.UIDIn(uids...),
		alertpage.Or(
			alertpage.OwnerUIDEQ(ownerUID),
			alertpage.PublicEQ(true),
		),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.AlertPage, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAlertPage(row))
	}
	return items, nil
}

func (r *alertPageRepo) ListUserAlertPages(ctx context.Context, namespaceUID, userUID string) ([]bo.AlertPage, error) {
	links, err := r.client.UserAlertPage.Query().Where(
		useralertpage.NamespaceUIDEQ(namespaceUID),
		useralertpage.UserUIDEQ(userUID),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if len(links) == 0 {
		return nil, nil
	}
	uids := make([]string, 0, len(links))
	for _, link := range links {
		uids = append(uids, link.AlertPageUID)
	}
	pages, err := r.ListSelectableByUIDs(ctx, namespaceUID, userUID, uids)
	if err != nil {
		return nil, err
	}
	byUID := make(map[string]bo.AlertPage, len(pages))
	for _, page := range pages {
		byUID[page.UID] = page
	}
	ordered := make([]bo.AlertPage, 0, len(links))
	for _, link := range links {
		if page, ok := byUID[link.AlertPageUID]; ok {
			ordered = append(ordered, page)
		}
	}
	return ordered, nil
}

func (r *alertPageRepo) SetUserAlertPages(ctx context.Context, namespaceUID, userUID string, pageUIDs []string) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.UserAlertPage.Delete().Where(
		useralertpage.NamespaceUIDEQ(namespaceUID),
		useralertpage.UserUIDEQ(userUID),
	).Exec(ctx); err != nil {
		return r.fail(err)
	}
	for _, pageUID := range pageUIDs {
		if _, err := tx.UserAlertPage.Create().
			SetUID(id.New()).
			SetNamespaceUID(namespaceUID).
			SetUserUID(userUID).
			SetAlertPageUID(pageUID).
			Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

func toAlertPage(row *ent.AlertPage) bo.AlertPage {
	return bo.AlertPage{
		UID:          row.UID,
		NamespaceUID: row.NamespaceUID,
		Name:         row.Name,
		OwnerUID:     row.OwnerUID,
		Labels:       row.Labels,
		Public:       row.Public,
		Remark:       row.Remark,
	}
}
