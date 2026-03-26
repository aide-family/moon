package impl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/plugin/cache"
	"github.com/aide-family/magicbox/safety"
	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/errors"
	klog "github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data"
	"github.com/aide-family/marksman/internal/data/impl/convert"
	"github.com/aide-family/marksman/internal/data/impl/do"
	"github.com/aide-family/marksman/internal/data/impl/query"
)

var alertEventTableCreateMu sync.Mutex

func NewAlertEventRepository(d *data.Data) (repository.AlertEvent, error) {
	query.SetDefault(d.DB())
	return &alertEventRepository{Data: d}, nil
}

type alertEventRepository struct {
	*data.Data
}

func (r *alertEventRepository) ensureAlertEventTable(ctx context.Context, tableName string) error {
	alertEventTableCreateMu.Lock()
	defer alertEventTableCreateMu.Unlock()
	if _, err := r.Cache().Get(ctx, cache.K(tableName)); err == nil && r.DB().Migrator().HasTable(tableName) {
		return nil
	}
	if r.DB().Migrator().HasTable(tableName) {
		return r.Cache().Set(ctx, cache.K(tableName), "", 0)
	}
	initModel := &do.AlertEvent{}
	baseName := initModel.TableName()
	if !r.DB().Migrator().HasTable(baseName) {
		if err := r.DB().Migrator().CreateTable(initModel); err != nil {
			return err
		}
	}
	if err := r.DB().Migrator().RenameTable(baseName, tableName); err != nil {
		return err
	}
	if err := r.Cache().Set(ctx, cache.K(tableName), "", 0); err != nil {
		klog.Context(ctx).Warnw("msg", "set cache for alert_event table failed", "error", err, "tableName", tableName)
	}
	return nil
}

func (r *alertEventRepository) SaveAlertEvent(ctx context.Context, ev *bo.AlertEventBo) (snowflake.ID, error) {
	snapshotID, err := r.findOrCreateEvaluatorSnapshot(ctx, ev.EvaluatorType, ev.EvaluatorSnapshotJSON)
	if err != nil {
		return 0, err
	}
	m := convert.ToAlertEventDo(ev, snapshotID)
	ns := contextx.GetNamespace(ctx)
	if ns.Int64() == 0 {
		ns = ev.NamespaceUID
	}
	tableName := do.GenAlertEventTableName(ns, ev.FiredAt)
	if err := r.ensureAlertEventTable(ctx, tableName); err != nil {
		return 0, err
	}

	table := query.AlertEvent.Table(tableName)
	wrappers := []gen.Condition{
		table.Fingerprint.Eq(ev.Fingerprint),
		table.NamespaceUID.Eq(ns.Int64()),
		table.StrategyUID.Eq(ev.StrategyUID.Int64()),
		table.LevelUID.Eq(ev.LevelUID.Int64()),
		table.Status.Eq(int32(enum.AlertEventStatus_ALERT_EVENT_STATUS_FIRING)),
	}
	info, err := table.WithContext(ctx).Where(wrappers...).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
		if err := table.WithContext(ctx).Create(m); err != nil {
			return 0, err
		}
		return m.ID, nil
	}
	updates := []field.AssignExpr{
		table.EvaluatorSnapshotID.Value(snapshotID.Int64()),
		table.UpdatedAt.Value(time.Now()),
		table.Value.Value(ev.Value),
		table.Labels.Value(safety.NewMap(ev.Labels)),
		table.Summary.Value(ev.Summary),
		table.Description.Value(ev.Description),
		table.Expr.Value(ev.Expr),
		table.StrategyName.Value(ev.StrategyName),
		table.StrategyGroupName.Value(ev.StrategyGroupName),
		table.LevelName.Value(ev.LevelName),
		table.DatasourceName.Value(ev.DatasourceName),
		table.BgColor.Value(ev.BgColor),
		table.DatasourceLevelName.Value(ev.DatasourceLevelName),
	}
	if _, err := table.WithContext(ctx).Where(table.ID.Eq(info.ID.Int64())).UpdateColumnSimple(updates...); err != nil {
		return 0, err
	}
	return info.ID, nil
}

// findOrCreateEvaluatorSnapshot returns evaluator_snapshot ID; dedupes by evaluator_type + content_hash.
func (r *alertEventRepository) findOrCreateEvaluatorSnapshot(ctx context.Context, evaluatorType, snapshotJSON string) (snowflake.ID, error) {
	hash := sha256.Sum256([]byte(snapshotJSON))
	contentHash := hex.EncodeToString(hash[:])
	e := query.Use(r.DB()).EvaluatorSnapshot
	snap, err := e.WithContext(ctx).Where(e.EvaluatorType.Eq(evaluatorType), e.ContentHash.Eq(contentHash)).First()
	if err == nil {
		return snap.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	newSnap := &do.EvaluatorSnapshot{
		EvaluatorType: evaluatorType,
		ContentHash:   contentHash,
		SnapshotJSON:  snapshotJSON,
	}
	if err = e.WithContext(ctx).Create(newSnap); err != nil {
		if isDuplicateKeyError(err) {
			snap, getErr := e.WithContext(ctx).Where(e.EvaluatorType.Eq(evaluatorType), e.ContentHash.Eq(contentHash)).First()
			if getErr != nil {
				return 0, getErr
			}
			return snap.ID, nil
		}
		return 0, err
	}
	return newSnap.ID, nil
}

func isDuplicateKeyError(err error) bool {
	return strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE")
}

func (r *alertEventRepository) GetAlertEvent(ctx context.Context, uid snowflake.ID) (*bo.AlertEventItemBo, error) {
	ns := contextx.GetNamespace(ctx)
	tableName := do.GenAlertEventTableName(ns, timex.TimeFromID(uid))
	if _, err := r.Cache().Get(ctx, cache.K(tableName)); err != nil && !r.DB().Migrator().HasTable(tableName) {
		return nil, merr.ErrorNotFound("alert event not found")
	}

	table := query.AlertEvent.Table(tableName)
	m, err := table.WithContext(ctx).Where(table.ID.Eq(uid.Int64())).Preload(field.Associations).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("alert event not found")
		}
		return nil, err
	}
	return convert.ToAlertEventItemBo(m), nil
}

func (r *alertEventRepository) GetAlertEventByFingerprint(ctx context.Context, uid snowflake.ID, fingerprint string) (*bo.AlertEventItemBo, error) {
	ns := contextx.GetNamespace(ctx)
	tableName := do.GenAlertEventTableName(ns, timex.TimeFromID(uid))
	if _, err := r.Cache().Get(ctx, cache.K(tableName)); err != nil && !r.DB().Migrator().HasTable(tableName) {
		return nil, merr.ErrorNotFound("alert event not found")
	}
	table := query.AlertEvent.Table(tableName)
	wrappers := []gen.Condition{
		table.NamespaceUID.Eq(ns.Int64()),
		table.ID.Eq(uid.Int64()),
		table.Fingerprint.Eq(fingerprint),
	}
	m, err := table.WithContext(ctx).Where(wrappers...).Preload(field.Associations).First()
	if err != nil {
		return nil, err
	}
	return convert.ToAlertEventItemBo(m), nil
}

func (r *alertEventRepository) ListRealtimeAlert(ctx context.Context, req *bo.ListRealtimeAlertBo, pageFilter *bo.AlertPageFilterBo) (*bo.PageResponseBo[*bo.AlertEventItemBo], error) {
	ns := contextx.GetNamespace(ctx)
	startAt := req.StartAt
	endAt := req.EndAt
	if startAt.IsZero() {
		startAt = time.Now().Add(-bo.ListRealtimeAlertTimeRangeDefault)
	}
	if endAt.IsZero() {
		endAt = time.Now()
	}

	tableNames := do.GenAlertEventTableNames(r.DB(), ns, startAt, endAt)
	if len(tableNames) == 0 {
		return bo.NewPageResponseBo[*bo.AlertEventItemBo](req.PageRequestBo, nil), nil
	}

	tables := make([]any, 0, len(tableNames))
	unionAllSQL := make([]string, 0, len(tableNames))
	for _, tableName := range tableNames {
		tables = append(tables, r.DB().Table(tableName))
		unionAllSQL = append(unionAllSQL, "?")
	}
	wrappers := r.DB().WithContext(ctx)
	if len(tableNames) > 1 {
		wrappers = wrappers.Table(fmt.Sprintf("(%s) as %s", strings.Join(unionAllSQL, " UNION ALL "), do.TableNameAlertEvent), tables...)
	} else {
		wrappers = wrappers.Table(fmt.Sprintf("%s as %s", tableNames[0], do.TableNameAlertEvent))
	}

	table := query.AlertEvent.As(do.TableNameAlertEvent)
	wrappers = wrappers.Where(table.NamespaceUID.Eq(ns.Int64())).
		Where(table.FiredAt.Gte(startAt)).
		Where(table.FiredAt.Lte(endAt))
	if req.Status != enum.AlertEventStatus_ALERT_EVENT_STATUS_UNKNOWN {
		wrappers = wrappers.Where(table.Status.Eq(int32(req.Status)))
	}
	wrappers, err := r.applyAlertPageFilter(ctx, wrappers, pageFilter)
	if err != nil {
		return nil, err
	}
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		wrappers = wrappers.Or(table.Summary.Like(keyword), table.Description.Like(keyword))
	}
	if len(req.StrategyGroupUids) > 0 {
		wrappers = wrappers.Where(table.StrategyGroupUID.In(req.StrategyGroupUids...))
	}
	if len(req.LevelUids) > 0 {
		wrappers = wrappers.Where(table.LevelUID.In(req.LevelUids...))
	}
	if len(req.StrategyUids) > 0 {
		wrappers = wrappers.Where(table.StrategyUID.In(req.StrategyUids...))
	}
	if len(req.DatasourceUids) > 0 {
		wrappers = wrappers.Where(table.DatasourceUID.In(req.DatasourceUids...))
	}
	var total int64
	if err := wrappers.Count(&total).Error; err != nil {
		return nil, err
	}
	req.WithTotal(total)
	if total == 0 {
		return bo.NewPageResponseBo[*bo.AlertEventItemBo](req.PageRequestBo, nil), nil
	}
	if req.Page > 0 && req.PageSize > 0 {
		wrappers = wrappers.Offset(req.Offset()).Limit(req.Limit())
	}
	wrappers = wrappers.Order(clause.OrderByColumn{Column: clause.Column{Name: table.FiredAt.ColumnName().String()}, Desc: true})
	var list []*do.AlertEvent
	if err := wrappers.Find(&list).Error; err != nil {
		return nil, err
	}
	items := make([]*bo.AlertEventItemBo, 0, len(list))
	for _, m := range list {
		items = append(items, convert.ToAlertEventItemBo(m))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func (r *alertEventRepository) InterveneAlert(ctx context.Context, req *bo.InterveneAlertBo) error {
	ns := contextx.GetNamespace(ctx)
	uid, by, byName := req.UID, req.IntervenedBy, req.IntervenedByName
	tableName := do.GenAlertEventTableName(ns, timex.TimeFromID(uid))
	if _, err := r.Cache().Get(ctx, cache.K(tableName)); err != nil && !r.DB().Migrator().HasTable(tableName) {
		return merr.ErrorNotFound("alert event not found")
	}
	table := query.AlertEvent.Table(tableName)
	info, err := r.GetAlertEvent(ctx, uid)
	if err != nil {
		return err
	}
	if info.IntervenedBy.Int64() > 0 && info.IntervenedByName != "" {
		return merr.ErrorConflict("alert event already intervened")
	}
	now := time.Now()
	wrappers := []gen.Condition{
		table.ID.Eq(uid.Int64()),
		table.Or(table.IntervenedBy.Eq(0), table.IntervenedByName.Eq("")),
		table.IntervenedAt.IsNull(),
	}
	_, err = table.WithContext(ctx).Where(wrappers...).UpdateColumnSimple(
		table.IntervenedAt.Value(now),
		table.IntervenedBy.Value(by.Int64()),
		table.IntervenedByName.Value(byName),
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *alertEventRepository) BatchInterveneAlert(ctx context.Context, req *bo.BatchInterveneAlertBo) error {
	ns := contextx.GetNamespace(ctx)
	now := time.Now()

	// Dedupe uids and group by shard table to avoid per-uid write statements.
	distinctUIDs := make(map[int64]struct{}, len(req.UIDs))
	tableUIDs := make(map[string][]int64)
	for _, uid := range req.UIDs {
		uidInt := uid.Int64()
		if uidInt <= 0 {
			continue
		}
		if _, ok := distinctUIDs[uidInt]; ok {
			continue
		}

		distinctUIDs[uidInt] = struct{}{}
		tableName := do.GenAlertEventTableName(ns, timex.TimeFromID(uid))
		if _, err := r.Cache().Get(ctx, cache.K(tableName)); err != nil && !r.DB().Migrator().HasTable(tableName) {
			klog.Context(ctx).Errorw("msg", "batch intervene alert failed", "error", err, "tableName", tableName)
			continue
		}
		tableUIDs[tableName] = append(tableUIDs[tableName], uidInt)
	}

	intervenedBy := req.IntervenedBy.Int64()
	intervenedByName := req.IntervenedByName

	for tableName, uids := range tableUIDs {
		if len(uids) == 0 {
			continue
		}

		table := query.AlertEvent.Table(tableName)
		_, err := table.WithContext(ctx).Where(table.ID.In(uids...), table.Or(table.IntervenedBy.Eq(0), table.IntervenedByName.Eq(""))).UpdateColumnSimple(
			table.IntervenedAt.Value(now),
			table.IntervenedBy.Value(intervenedBy),
			table.IntervenedByName.Value(intervenedByName),
		)
		if err != nil {
			klog.Context(ctx).Errorw("msg", "batch intervene alert failed", "error", err, "tableName", tableName, "uids", uids)
			continue
		}
	}
	return nil
}

func (r *alertEventRepository) SuppressAlert(ctx context.Context, req *bo.SuppressAlertBo) error {
	uid, until, by, byName, reason := req.UID, req.SuppressUntilAt, req.SuppressedBy, req.SuppressedByName, req.SuppressedReason
	ns := contextx.GetNamespace(ctx)
	tableName := do.GenAlertEventTableName(ns, timex.TimeFromID(uid))
	if _, err := r.Cache().Get(ctx, cache.K(tableName)); err != nil && !r.DB().Migrator().HasTable(tableName) {
		return merr.ErrorNotFound("alert event not found")
	}
	table := query.AlertEvent.Table(tableName)
	info, err := table.WithContext(ctx).Where(table.ID.Eq(uid.Int64())).UpdateColumnSimple(
		table.SuppressedUntilAt.Value(until),
		table.SuppressedBy.Value(by.Int64()),
		table.SuppressedByName.Value(byName),
		table.SuppressedReason.Value(reason),
	)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("alert event not found")
	}
	return nil
}

func (r *alertEventRepository) RecoverAlert(ctx context.Context, req *bo.RecoverAlertBo) error {
	uid, by, byName, reason := req.UID, req.RecoveredBy, req.RecoveredByName, req.RecoveredReason
	ns := contextx.GetNamespace(ctx)
	tableName := do.GenAlertEventTableName(ns, timex.TimeFromID(uid))
	if _, err := r.Cache().Get(ctx, cache.K(tableName)); err != nil && !r.DB().Migrator().HasTable(tableName) {
		return merr.ErrorNotFound("alert event not found")
	}
	table := query.AlertEvent.Table(tableName)
	now := time.Now()
	info, err := table.WithContext(ctx).Where(table.ID.Eq(uid.Int64())).UpdateColumnSimple(
		table.Status.Value(int32(enum.AlertEventStatus_ALERT_EVENT_STATUS_RECOVERED_BY_MANUAL)),
		table.RecoveredAt.Value(now),
		table.RecoveredBy.Value(by.Int64()),
		table.RecoveredByName.Value(byName),
		table.RecoveredReason.Value(reason),
	)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("alert event not found")
	}
	return nil
}

func (r *alertEventRepository) AutoRecoverAlert(ctx context.Context, uid snowflake.ID) error {
	ns := contextx.GetNamespace(ctx)
	tableName := do.GenAlertEventTableName(ns, timex.TimeFromID(uid))
	if _, err := r.Cache().Get(ctx, cache.K(tableName)); err != nil && !r.DB().Migrator().HasTable(tableName) {
		return merr.ErrorNotFound("alert event not found")
	}
	table := query.AlertEvent.Table(tableName)
	_, err := table.WithContext(ctx).Where(table.ID.Eq(uid.Int64()), table.Status.Eq(int32(enum.AlertEventStatus_ALERT_EVENT_STATUS_FIRING))).UpdateColumnSimple(
		table.Status.Value(int32(enum.AlertEventStatus_ALERT_EVENT_STATUS_RECOVERED)),
		table.RecoveredAt.Value(time.Now()),
	)
	if err != nil {
		return err
	}

	return nil
}

// buildAlertEventUnion builds a GORM DB scope over UNION of alert_events shard tables for ns and [startAt,endAt].
// Caller must add Where conditions and run Count/Find/Scan. Table alias is do.TableNameAlertEvent.
func (r *alertEventRepository) buildAlertEventUnion(ctx context.Context, ns snowflake.ID, startAt, endAt time.Time) *gorm.DB {
	tableNames := do.GenAlertEventTableNames(r.DB(), ns, startAt, endAt)
	if len(tableNames) == 0 {
		return r.DB().WithContext(ctx).Table(do.TableNameAlertEvent).Where("1 = 0")
	}
	unionAllSQL := make([]string, 0, len(tableNames))
	for range tableNames {
		unionAllSQL = append(unionAllSQL, "?")
	}
	wrappers := r.DB().WithContext(ctx)
	if len(tableNames) > 1 {
		wrappers = wrappers.Table(fmt.Sprintf("(%s) as %s", strings.Join(unionAllSQL, " UNION ALL "), do.TableNameAlertEvent), r.tablesFromNames(tableNames)...)
	} else {
		wrappers = wrappers.Table(fmt.Sprintf("%s as %s", tableNames[0], do.TableNameAlertEvent))
	}
	table := query.AlertEvent.As(do.TableNameAlertEvent)
	return wrappers.Where(table.NamespaceUID.Eq(ns.Int64())).
		Where(table.FiredAt.Gte(startAt)).
		Where(table.FiredAt.Lte(endAt))
}

func (r *alertEventRepository) tablesFromNames(names []string) []any {
	out := make([]any, 0, len(names))
	for _, n := range names {
		out = append(out, r.DB().Table(n))
	}
	return out
}

func (r *alertEventRepository) datasourceUIDsByLevelUIDs(ctx context.Context, levelUIDs []int64) ([]int64, error) {
	if len(levelUIDs) == 0 {
		return nil, nil
	}
	d := query.Datasource
	list, err := d.WithContext(ctx).Where(
		d.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		d.LevelUID.In(levelUIDs...),
	).Select(d.ID).Find()
	if err != nil {
		return nil, err
	}
	uidSet := make(map[int64]struct{}, len(list))
	for _, item := range list {
		uidSet[item.ID.Int64()] = struct{}{}
	}
	out := make([]int64, 0, len(uidSet))
	for id := range uidSet {
		out = append(out, id)
	}
	return out, nil
}

func (r *alertEventRepository) applyAlertPageFilter(
	ctx context.Context,
	db *gorm.DB,
	pageFilter *bo.AlertPageFilterBo,
) (*gorm.DB, error) {
	table := query.AlertEvent.As(do.TableNameAlertEvent)
	if pageFilter == nil {
		return db, nil
	}
	if len(pageFilter.StrategyUIDs) > 0 {
		db = db.Where(table.StrategyUID.In(pageFilter.StrategyUIDs...))
	}
	if len(pageFilter.LevelUIDs) > 0 {
		db = db.Where(table.LevelUID.In(pageFilter.LevelUIDs...))
	}
	if len(pageFilter.StrategyGroupUIDs) > 0 {
		db = db.Where(table.StrategyGroupUID.In(pageFilter.StrategyGroupUIDs...))
	}
	if len(pageFilter.DatasourceUIDs) > 0 {
		db = db.Where(table.DatasourceUID.In(pageFilter.DatasourceUIDs...))
	}
	if len(pageFilter.DatasourceLevelUIDs) > 0 {
		datasourceUIDs, err := r.datasourceUIDsByLevelUIDs(ctx, pageFilter.DatasourceLevelUIDs)
		if err != nil {
			return nil, err
		}
		if len(datasourceUIDs) == 0 {
			return db.Where("1 = 0"), nil
		}
		db = db.Where(table.DatasourceUID.In(datasourceUIDs...))
	}
	return db, nil
}

func (r *alertEventRepository) applyActiveFilter(ctx context.Context, db *gorm.DB, pageFilter *bo.AlertPageFilterBo) (*gorm.DB, error) {
	table := query.AlertEvent.As(do.TableNameAlertEvent)
	// Active alerts are persisted with FIRING status.
	db = db.Where(table.Status.Eq(int32(enum.AlertEventStatus_ALERT_EVENT_STATUS_FIRING)))
	return r.applyAlertPageFilter(ctx, db, pageFilter)
}

func (r *alertEventRepository) CountActiveAlerts(ctx context.Context, startAt, endAt time.Time, pageFilter *bo.AlertPageFilterBo) (int64, error) {
	ns := contextx.GetNamespace(ctx)
	if ns.Int64() == 0 {
		return 0, nil
	}
	db := r.buildAlertEventUnion(ctx, ns, startAt, endAt)
	db, err := r.applyActiveFilter(ctx, db, pageFilter)
	if err != nil {
		return 0, err
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

type levelCountRow struct {
	LevelUID int64 `gorm:"column:level_uid"`
	Count    int64 `gorm:"column:count"`
}

func (r *alertEventRepository) CountActiveAlertsByLevel(ctx context.Context, startAt, endAt time.Time, pageFilter *bo.AlertPageFilterBo) ([]bo.LevelCountBo, error) {
	ns := contextx.GetNamespace(ctx)
	if ns.Int64() == 0 {
		return nil, nil
	}
	db := r.buildAlertEventUnion(ctx, ns, startAt, endAt)
	db, err := r.applyActiveFilter(ctx, db, pageFilter)
	if err != nil {
		return nil, err
	}
	var rows []levelCountRow
	if err := db.Select("level_uid, count(*) as count").Group("level_uid").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]bo.LevelCountBo, 0, len(rows))
	for _, row := range rows {
		out = append(out, bo.LevelCountBo{LevelUID: snowflake.ParseInt64(row.LevelUID), Count: row.Count})
	}
	return out, nil
}

func (r *alertEventRepository) CountRecoveredAlertsSince(ctx context.Context, since time.Time) (int64, error) {
	ns := contextx.GetNamespace(ctx)
	if ns.Int64() == 0 {
		return 0, nil
	}
	startAt := since.Add(-bo.ListRealtimeAlertTimeRangeDefault)
	endAt := time.Now()
	db := r.buildAlertEventUnion(ctx, ns, startAt, endAt)
	table := query.AlertEvent.As(do.TableNameAlertEvent)
	db = db.Where(table.Status.In(
		int32(enum.AlertEventStatus_ALERT_EVENT_STATUS_RECOVERED),
		int32(enum.AlertEventStatus_ALERT_EVENT_STATUS_RECOVERED_BY_MANUAL),
	)).Where(table.RecoveredAt.Gte(since))
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
