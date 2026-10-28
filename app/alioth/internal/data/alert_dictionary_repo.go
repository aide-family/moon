package data

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertdictionary"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertrulelevel"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *alertDictionaryRepo) CreateAlertDictionary(ctx context.Context, in bo.AlertDictionary) (*bo.AlertDictionary, error) {
	kind, err := dictionaryType(in.Type)
	if err != nil {
		return nil, err
	}
	row, err := r.client.AlertDictionary.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetColor(in.Color).
		SetType(kind).
		SetStatus(alertdictionary.StatusACTIVE).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toAlertDictionary(row)
	return &item, nil
}

func (r *alertDictionaryRepo) UpdateAlertDictionary(ctx context.Context, in bo.AlertDictionary) (*bo.AlertDictionary, error) {
	n, err := r.client.AlertDictionary.Update().Where(
		alertdictionary.NamespaceUIDEQ(in.NamespaceUID),
		alertdictionary.UIDEQ(in.UID),
		alertdictionary.DeletedAtIsNil(),
	).SetName(in.Name).SetColor(in.Color).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetAlertDictionary(ctx, in.NamespaceUID, in.UID)
}

func (r *alertDictionaryRepo) UpdateAlertDictionaryStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.AlertDictionary, error) {
	value := alertdictionary.StatusACTIVE
	if status == "DISABLED" {
		value = alertdictionary.StatusDISABLED
	} else if status != "ACTIVE" {
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.AlertDictionary.Update().Where(
		alertdictionary.NamespaceUIDEQ(namespaceUID),
		alertdictionary.UIDEQ(uid),
		alertdictionary.DeletedAtIsNil(),
	).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetAlertDictionary(ctx, namespaceUID, uid)
}

func (r *alertDictionaryRepo) DeleteAlertDictionary(ctx context.Context, namespaceUID, uid string) error {
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
	n, err := r.client.AlertDictionary.Update().Where(
		alertdictionary.NamespaceUIDEQ(namespaceUID),
		alertdictionary.UIDEQ(uid),
		alertdictionary.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

func (r *alertDictionaryRepo) GetAlertDictionary(ctx context.Context, namespaceUID, uid string) (*bo.AlertDictionary, error) {
	row, err := r.client.AlertDictionary.Query().Where(
		alertdictionary.NamespaceUIDEQ(namespaceUID),
		alertdictionary.UIDEQ(uid),
		alertdictionary.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toAlertDictionary(row)
	return &item, nil
}

func (r *alertDictionaryRepo) ListAlertDictionaries(ctx context.Context, namespaceUID string, page, pageSize int, kind, name string) ([]bo.AlertDictionary, int, error) {
	query := r.client.AlertDictionary.Query().Where(alertdictionary.NamespaceUIDEQ(namespaceUID), alertdictionary.DeletedAtIsNil())
	if name != "" {
		query = query.Where(alertdictionary.NameContainsFold(name))
	}
	if kind != "" {
		parsed, err := dictionaryType(kind)
		if err != nil {
			return nil, 0, err
		}
		query = query.Where(alertdictionary.TypeEQ(parsed))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(alertdictionary.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.AlertDictionary, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAlertDictionary(row))
	}
	return items, total, nil
}

func toAlertDictionary(row *ent.AlertDictionary) bo.AlertDictionary {
	return bo.AlertDictionary{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name,
		Color: row.Color, Type: string(row.Type), Status: string(row.Status),
	}
}

func dictionaryType(kind string) (alertdictionary.Type, error) {
	switch alertdictionary.Type(kind) {
	case alertdictionary.TypeALERT, alertdictionary.TypeDATASOURCE:
		return alertdictionary.Type(kind), nil
	default:
		return "", merr.Invalid("invalid request")
	}
}
