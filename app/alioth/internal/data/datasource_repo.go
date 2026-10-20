package data

import (
	"context"
	"fmt"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertdictionary"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/datasource"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	labelsutil "github.com/aide-family/Alioth/pkg/labels"
	"github.com/aide-family/Alioth/pkg/merr"
)

func (r *datasourceRepo) CreateDatasource(ctx context.Context, in bo.Datasource) (*bo.Datasource, error) {
	kind, driver, status, err := datasourceEnums(in.Type, in.Driver, "ACTIVE")
	if err != nil {
		return nil, err
	}
	password, err := r.box.encrypt(in.Password)
	if err != nil {
		return nil, err
	}
	secret, err := r.box.encrypt(in.Secret)
	if err != nil {
		return nil, err
	}
	row, err := r.client.Datasource.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetType(kind).
		SetDriver(driver).
		SetEndpoint(in.Endpoint).
		SetIndexPattern(in.IndexPattern).
		SetTimeField(in.TimeField).
		SetUsername(in.Username).
		SetPasswordCipher(password).
		SetSecretCipher(secret).
		SetStatus(status).
		SetLevelUID(in.LevelUID).
		SetLabels(in.Labels).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.present(ctx, row, true)
}

func (r *datasourceRepo) UpdateDatasource(ctx context.Context, in bo.Datasource) (*bo.Datasource, error) {
	row, err := r.client.Datasource.Query().Where(
		datasource.NamespaceUIDEQ(in.NamespaceUID),
		datasource.UIDEQ(in.UID),
		datasource.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	kind, driver, _, err := datasourceEnums(in.Type, in.Driver, string(row.Status))
	if err != nil {
		return nil, err
	}
	password := row.PasswordCipher
	if in.Password != "" {
		password, err = r.box.encrypt(in.Password)
		if err != nil {
			return nil, err
		}
	}
	secret := row.SecretCipher
	if in.Secret != "" {
		secret, err = r.box.encrypt(in.Secret)
		if err != nil {
			return nil, err
		}
	}
	updated, err := r.client.Datasource.UpdateOne(row).
		SetName(in.Name).
		SetType(kind).
		SetDriver(driver).
		SetEndpoint(in.Endpoint).
		SetIndexPattern(in.IndexPattern).
		SetTimeField(in.TimeField).
		SetUsername(in.Username).
		SetPasswordCipher(password).
		SetSecretCipher(secret).
		SetLevelUID(in.LevelUID).
		SetLabels(in.Labels).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.present(ctx, updated, true)
}

func (r *datasourceRepo) UpdateDatasourceStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.Datasource, error) {
	value := datasource.StatusACTIVE
	if status == "DISABLED" {
		value = datasource.StatusDISABLED
	} else if status != "ACTIVE" {
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.Datasource.Update().Where(
		datasource.NamespaceUIDEQ(namespaceUID),
		datasource.UIDEQ(uid),
		datasource.DeletedAtIsNil(),
	).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetDatasource(ctx, namespaceUID, uid)
}

func (r *datasourceRepo) DeleteDatasource(ctx context.Context, namespaceUID, uid string) error {
	n, err := r.client.Datasource.Update().Where(
		datasource.NamespaceUIDEQ(namespaceUID),
		datasource.UIDEQ(uid),
		datasource.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

func (r *datasourceRepo) GetDatasource(ctx context.Context, namespaceUID, uid string) (*bo.Datasource, error) {
	row, err := r.client.Datasource.Query().Where(
		datasource.NamespaceUIDEQ(namespaceUID),
		datasource.UIDEQ(uid),
		datasource.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.present(ctx, row, true)
}

func (r *datasourceRepo) ListDatasources(ctx context.Context, namespaceUID string, page, pageSize int, name, kind string) ([]bo.Datasource, int, error) {
	query := r.client.Datasource.Query().Where(datasource.NamespaceUIDEQ(namespaceUID), datasource.DeletedAtIsNil())
	if name != "" {
		query = query.Where(datasource.NameContainsFold(name))
	}
	if kind != "" {
		if kind != "METRICS" && kind != "LOGS" && kind != "TRACE" {
			return nil, 0, merr.Invalid("invalid request")
		}
		query = query.Where(datasource.TypeEQ(datasource.Type(kind)))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(datasource.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	return r.presentMany(ctx, rows, false), total, nil
}

func (r *datasourceRepo) ListDatasourcesByMatchers(ctx context.Context, namespaceUID, kind, matchersJSON string) ([]bo.Datasource, error) {
	if kind != "METRICS" && kind != "LOGS" && kind != "TRACE" {
		return nil, merr.Invalid("invalid request")
	}
	rows, err := r.client.Datasource.Query().Where(
		datasource.NamespaceUIDEQ(namespaceUID),
		datasource.DeletedAtIsNil(),
		datasource.StatusEQ(datasource.StatusACTIVE),
		datasource.TypeEQ(datasource.Type(kind)),
	).Order(ent.Asc(datasource.FieldUID)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.Datasource, 0, len(rows))
	for _, row := range rows {
		ok, err := labelsutil.MatchersJSONContained(matchersJSON, row.Labels)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		item, err := r.present(ctx, row, true)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

func (r *datasourceRepo) SelectDatasources(ctx context.Context, namespaceUID string, limit int, lastUID, kind string) ([]bo.SelectItem, error) {
	query := r.client.Datasource.Query().Where(datasource.NamespaceUIDEQ(namespaceUID), datasource.DeletedAtIsNil())
	if lastUID != "" {
		query = query.Where(datasource.UIDGT(lastUID))
	}
	if kind == "METRICS" || kind == "LOGS" || kind == "TRACE" {
		query = query.Where(datasource.TypeEQ(datasource.Type(kind)))
	}
	rows, err := query.Order(ent.Asc(datasource.FieldUID)).Limit(limit).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.SelectItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, bo.SelectItem{UID: row.UID, Name: row.Name})
	}
	return items, nil
}

func (r *datasourceRepo) present(ctx context.Context, row *ent.Datasource, withSecret bool) (*bo.Datasource, error) {
	names := r.levelNames(ctx, []string{row.LevelUID})
	item := toDatasource(row, names[row.LevelUID], withSecret, r)
	return &item, nil
}

func (r *datasourceRepo) presentMany(ctx context.Context, rows []*ent.Datasource, withSecret bool) []bo.Datasource {
	uids := make([]string, 0, len(rows))
	for _, row := range rows {
		uids = append(uids, row.LevelUID)
	}
	names := r.levelNames(ctx, uids)
	items := make([]bo.Datasource, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDatasource(row, names[row.LevelUID], withSecret, r))
	}
	return items
}

func toDatasource(row *ent.Datasource, levelName string, withSecret bool, r *datasourceRepo) bo.Datasource {
	item := bo.Datasource{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name,
		Type: string(row.Type), Driver: string(row.Driver), Endpoint: row.Endpoint,
		IndexPattern: row.IndexPattern, TimeField: row.TimeField,
		Username: row.Username, SecretSet: row.PasswordCipher != "" || row.SecretCipher != "",
		Status: string(row.Status), LevelUID: row.LevelUID, LevelName: levelName,
		Labels: row.Labels, Remark: row.Remark,
	}
	if !withSecret {
		return item
	}
	password, err := r.box.decrypt(row.PasswordCipher)
	if err != nil {
		r.log.Error(fmt.Sprintf("datasource %s decrypt", row.UID))
		password = ""
	}
	secret, err := r.box.decrypt(row.SecretCipher)
	if err != nil {
		r.log.Error(fmt.Sprintf("datasource %s decrypt", row.UID))
		secret = ""
	}
	item.Password = password
	item.Secret = secret
	return item
}

func (r *datasourceRepo) levelNames(ctx context.Context, uids []string) map[string]string {
	names := map[string]string{}
	clean := make([]string, 0, len(uids))
	for _, uid := range uids {
		if uid != "" {
			clean = append(clean, uid)
		}
	}
	if len(clean) == 0 {
		return names
	}
	rows, err := r.client.AlertDictionary.Query().Where(alertdictionary.UIDIn(clean...), alertdictionary.DeletedAtIsNil()).All(ctx)
	if err != nil {
		r.log.Error(fmt.Sprintf("level names: %v", err))
		return names
	}
	for _, row := range rows {
		names[row.UID] = row.Name
	}
	return names
}

func datasourceEnums(kind, driver, status string) (datasource.Type, datasource.Driver, datasource.Status, error) {
	parsedType := datasource.Type(kind)
	switch parsedType {
	case datasource.TypeMETRICS, datasource.TypeLOGS, datasource.TypeTRACE:
	default:
		return "", "", "", merr.Invalid("invalid request")
	}
	parsedDriver := datasource.Driver(driver)
	switch parsedDriver {
	case datasource.DriverMETRICS_PROMETHEUS, datasource.DriverMETRICS_VICTORIA_METRICS, datasource.DriverLOGS_ELASTICSEARCH, datasource.DriverLOGS_LOKI, datasource.DriverLOGS_VLOGS, datasource.DriverTRACE_JAEGER:
	default:
		return "", "", "", merr.Invalid("invalid request")
	}
	parsedStatus := datasource.StatusACTIVE
	if status == "DISABLED" {
		parsedStatus = datasource.StatusDISABLED
	} else if status != "ACTIVE" {
		return "", "", "", merr.Invalid("invalid request")
	}
	return parsedType, parsedDriver, parsedStatus, nil
}
