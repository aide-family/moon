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
		table.FiredAt.Value(info.FiredAt),
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
	levelInfo := r.levelInfoByUID(ctx, m.NamespaceUID, m.LevelUID)
	return convert.ToAlertEventItemBo(m, levelInfo.Name, levelInfo.BgColor), nil
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
	levelInfo := r.levelInfoByUID(ctx, m.NamespaceUID, m.LevelUID)
	return convert.ToAlertEventItemBo(m, levelInfo.Name, levelInfo.BgColor), nil
}

type levelInfo struct {
	Name    string
	BgColor string
}

func (r *alertEventRepository) levelInfoByUID(ctx context.Context, namespaceUID, levelUID snowflake.ID) levelInfo {
	l := query.Level
	lev, err := l.WithContext(ctx).Where(
		l.NamespaceUID.Eq(namespaceUID.Int64()),
		l.ID.Eq(levelUID.Int64()),
	).First()
	if err != nil {
		return levelInfo{}
	}
	return levelInfo{
		Name:    lev.Name,
		BgColor: lev.BgColor,
	}
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
	if pageFilter != nil {
		if len(pageFilter.StrategyUIDs) > 0 {
			wrappers = wrappers.Where(table.StrategyUID.In(pageFilter.StrategyUIDs...))
		}
		if len(pageFilter.LevelUIDs) > 0 {
			wrappers = wrappers.Where(table.LevelUID.In(pageFilter.LevelUIDs...))
		}
		if len(pageFilter.StrategyGroupUIDs) > 0 {
			wrappers = wrappers.Where(table.StrategyGroupUID.In(pageFilter.StrategyGroupUIDs...))
		}
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
	levelInfos := r.levelInfosForEvents(ctx, ns, list)
	items := make([]*bo.AlertEventItemBo, 0, len(list))
	for _, m := range list {
		info := levelInfos[m.LevelUID.Int64()]
		items = append(items, convert.ToAlertEventItemBo(m, info.Name, info.BgColor))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func (r *alertEventRepository) levelInfosForEvents(ctx context.Context, namespaceUID snowflake.ID, events []*do.AlertEvent) map[int64]levelInfo {
	uidSet := make(map[int64]struct{})
	for _, ev := range events {
		uidSet[ev.LevelUID.Int64()] = struct{}{}
	}
	if len(uidSet) == 0 {
		return nil
	}
	uids := make([]int64, 0, len(uidSet))
	for id := range uidSet {
		uids = append(uids, id)
	}
	l := query.Level
	levels, err := l.WithContext(ctx).Where(
		l.NamespaceUID.Eq(namespaceUID.Int64()),
		l.ID.In(uids...),
	).Select(l.ID, l.Name, l.BgColor).Find()
	if err != nil {
		return nil
	}
	out := make(map[int64]levelInfo, len(levels))
	for _, lev := range levels {
		out[lev.ID.Int64()] = levelInfo{
			Name:    lev.Name,
			BgColor: lev.BgColor,
		}
	}
	return out
}

func (r *alertEventRepository) InterveneAlert(ctx context.Context, req *bo.InterveneAlertBo) error {
	ns := contextx.GetNamespace(ctx)
	uid, by, byName := req.UID, req.IntervenedBy, req.IntervenedByName
	tableName := do.GenAlertEventTableName(ns, timex.TimeFromID(uid))
	if _, err := r.Cache().Get(ctx, cache.K(tableName)); err != nil && !r.DB().Migrator().HasTable(tableName) {
		return merr.ErrorNotFound("alert event not found")
	}
	table := query.AlertEvent.Table(tableName)
	now := time.Now()
	info, err := table.WithContext(ctx).Where(table.ID.Eq(uid.Int64())).UpdateColumnSimple(
		table.IntervenedAt.Value(now),
		table.IntervenedBy.Value(by.Int64()),
		table.IntervenedByName.Value(byName),
	)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("alert event not found")
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

func (r *alertEventRepository) applyActiveFilter(db *gorm.DB, pageFilter *bo.AlertPageFilterBo) *gorm.DB {
	table := query.AlertEvent.As(do.TableNameAlertEvent)
	// Active alerts are persisted with FIRING status.
	db = db.Where(table.Status.Eq(int32(enum.AlertEventStatus_ALERT_EVENT_STATUS_FIRING)))
	if pageFilter != nil {
		if len(pageFilter.StrategyUIDs) > 0 {
			db = db.Where(table.StrategyUID.In(pageFilter.StrategyUIDs...))
		}
		if len(pageFilter.LevelUIDs) > 0 {
			db = db.Where(table.LevelUID.In(pageFilter.LevelUIDs...))
		}
		if len(pageFilter.StrategyGroupUIDs) > 0 {
			db = db.Where(table.StrategyGroupUID.In(pageFilter.StrategyGroupUIDs...))
		}
	}
	return db
}

func (r *alertEventRepository) CountActiveAlerts(ctx context.Context, startAt, endAt time.Time, pageFilter *bo.AlertPageFilterBo) (int64, error) {
	ns := contextx.GetNamespace(ctx)
	if ns.Int64() == 0 {
		return 0, nil
	}
	db := r.buildAlertEventUnion(ctx, ns, startAt, endAt)
	db = r.applyActiveFilter(db, pageFilter)
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
	db = r.applyActiveFilter(db, pageFilter)
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
