package data

import (
	"context"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertlevel"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrulelevel"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/pkg/merr"
)

func (r *alertLevelRepo) CreateAlertLevel(ctx context.Context, in bo.AlertLevel) (*bo.AlertLevel, error) {
	kind, err := levelType(in.Type)
	if err != nil {
		return nil, err
	}
	row, err := r.client.AlertLevel.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetColor(in.Color).
		SetType(kind).
		SetStatus(alertlevel.StatusACTIVE).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toAlertLevel(row)
	return &item, nil
}

func (r *alertLevelRepo) UpdateAlertLevel(ctx context.Context, in bo.AlertLevel) (*bo.AlertLevel, error) {
	n, err := r.client.AlertLevel.Update().Where(
		alertlevel.NamespaceUIDEQ(in.NamespaceUID),
		alertlevel.UIDEQ(in.UID),
		alertlevel.DeletedAtIsNil(),
	).SetName(in.Name).SetColor(in.Color).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetAlertLevel(ctx, in.NamespaceUID, in.UID)
}

func (r *alertLevelRepo) UpdateAlertLevelStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.AlertLevel, error) {
	value := alertlevel.StatusACTIVE
	if status == "DISABLED" {
		value = alertlevel.StatusDISABLED
	} else if status != "ACTIVE" {
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.AlertLevel.Update().Where(
		alertlevel.NamespaceUIDEQ(namespaceUID),
		alertlevel.UIDEQ(uid),
		alertlevel.DeletedAtIsNil(),
	).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetAlertLevel(ctx, namespaceUID, uid)
}

func (r *alertLevelRepo) DeleteAlertLevel(ctx context.Context, namespaceUID, uid string) error {
	used, err := r.client.AlertRuleLevel.Query().Where(
		alertrulelevel.NamespaceUIDEQ(namespaceUID),
		alertrulelevel.LevelUIDEQ(uid),
		alertrulelevel.DeletedAtIsNil(),
	).Exist(ctx)
	if err != nil {
		return r.fail(err)
	}
	if used {
		return merr.Conflict("level is in use")
	}
	n, err := r.client.AlertLevel.Update().Where(
		alertlevel.NamespaceUIDEQ(namespaceUID),
		alertlevel.UIDEQ(uid),
		alertlevel.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

func (r *alertLevelRepo) GetAlertLevel(ctx context.Context, namespaceUID, uid string) (*bo.AlertLevel, error) {
	row, err := r.client.AlertLevel.Query().Where(
		alertlevel.NamespaceUIDEQ(namespaceUID),
		alertlevel.UIDEQ(uid),
		alertlevel.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toAlertLevel(row)
	return &item, nil
}

func (r *alertLevelRepo) ListAlertLevels(ctx context.Context, namespaceUID string, page, pageSize int, kind, name string) ([]bo.AlertLevel, int, error) {
	query := r.client.AlertLevel.Query().Where(alertlevel.NamespaceUIDEQ(namespaceUID), alertlevel.DeletedAtIsNil())
	if name != "" {
		query = query.Where(alertlevel.NameContainsFold(name))
	}
	if kind != "" {
		parsed, err := levelType(kind)
		if err != nil {
			return nil, 0, err
		}
		query = query.Where(alertlevel.TypeEQ(parsed))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(alertlevel.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.AlertLevel, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAlertLevel(row))
	}
	return items, total, nil
}

func toAlertLevel(row *ent.AlertLevel) bo.AlertLevel {
	return bo.AlertLevel{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name,
		Color: row.Color, Type: string(row.Type), Status: string(row.Status),
	}
}

func levelType(kind string) (alertlevel.Type, error) {
	switch alertlevel.Type(kind) {
	case alertlevel.TypeALERT, alertlevel.TypeDATASOURCE:
		return alertlevel.Type(kind), nil
	default:
		return "", merr.Invalid("invalid request")
	}
}
