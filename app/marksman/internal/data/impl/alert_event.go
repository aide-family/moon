package impl

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/plugin/cache"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/errors"
	klog "github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"

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
	if _, err := r.Cache().Get(ctx, cache.K(tableName)); err == nil && r.DB().Migrator().HasTable(tableName) {
		return nil
	}
	if r.DB().Migrator().HasTable(tableName) {
		_ = r.Cache().Set(ctx, cache.K(tableName), "", 0)
		return nil
	}
	alertEventTableCreateMu.Lock()
	defer alertEventTableCreateMu.Unlock()
	if r.DB().Migrator().HasTable(tableName) {
		_ = r.Cache().Set(ctx, cache.K(tableName), "", 0)
		return nil
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

func (r *alertEventRepository) CreateAlertEvent(ctx context.Context, ev *bo.AlertEventBo, strategyGroupUID snowflake.ID) error {
	m := convert.ToAlertEventDo(ev, strategyGroupUID)
	ns := contextx.GetNamespace(ctx)
	if ns.Int64() == 0 {
		ns = ev.NamespaceUID
	}
	tableName := do.GenAlertEventTableName(ns, ev.FiredAt)
	if err := r.ensureAlertEventTable(ctx, tableName); err != nil {
		return err
	}
	bizQuery := query.Use(r.DB().Table(tableName))
	alertEvent := bizQuery.AlertEvent
	return alertEvent.WithContext(ctx).Create(m)
}

func (r *alertEventRepository) GetAlertEvent(ctx context.Context, uid snowflake.ID) (*bo.AlertEventItemBo, error) {
	ns := contextx.GetNamespace(ctx)
	tableName := do.GenAlertEventTableName(ns, do.AlertEventTimeFromID(uid))
	if _, err := r.Cache().Get(ctx, cache.K(tableName)); err != nil && !r.DB().Migrator().HasTable(tableName) {
		return nil, merr.ErrorNotFound("alert event not found")
	}
	bizQuery := query.Use(r.DB().Table(tableName))
	e := bizQuery.AlertEvent
	table := e.As(tableName)
	m, err := e.WithContext(ctx).Where(table.ID.Eq(uid.Int64())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("alert event not found")
		}
		return nil, err
	}
	levelName := r.levelNameByUID(ctx, m.NamespaceUID, m.LevelUID)
	return convert.ToAlertEventItemBo(m, levelName), nil
}

func (r *alertEventRepository) levelNameByUID(ctx context.Context, namespaceUID, levelUID snowflake.ID) string {
	l := query.Level
	lev, err := l.WithContext(ctx).Where(
		l.NamespaceUID.Eq(namespaceUID.Int64()),
		l.ID.Eq(levelUID.Int64()),
	).First()
	if err != nil {
		return ""
	}
	return lev.Name
}

func (r *alertEventRepository) ListRealtimeAlert(ctx context.Context, req *bo.ListRealtimeAlertBo, pageFilter *bo.AlertPageFilterBo) (*bo.PageResponseBo[*bo.AlertEventItemBo], error) {
	ns := contextx.GetNamespace(ctx)
	endAt := time.Now()
	startAt := endAt.AddDate(0, 0, -14)

	tableNames := do.GenAlertEventTableNames(r.DB(), ns, startAt, endAt)
	if len(tableNames) == 0 {
		return bo.NewPageResponseBo[*bo.AlertEventItemBo](req.PageRequestBo, nil), nil
	}

	var tx *gorm.DB
	if len(tableNames) == 1 {
		tx = r.DB().WithContext(ctx).Table(tableNames[0]).Model(&do.AlertEvent{}).Where("namespace_uid = ?", ns.Int64())
	} else {
		tables := make([]any, 0, len(tableNames))
		unionParts := make([]string, 0, len(tableNames))
		for _, tableName := range tableNames {
			tables = append(tables, r.DB().Table(tableName))
			unionParts = append(unionParts, "?")
		}
		unionSQL := fmt.Sprintf("(%s) as %s", strings.Join(unionParts, " UNION ALL "), do.TableNameAlertEvent)
		tx = r.DB().WithContext(ctx).Table(unionSQL, tables...).Model(&do.AlertEvent{}).
			Where("namespace_uid = ?", ns.Int64()).
			Where("fired_at >= ?", startAt).
			Where("fired_at <= ?", endAt)
	}

	if req.Status != apiv1.AlertEventStatus_ALERT_EVENT_STATUS_UNKNOWN {
		tx = tx.Where("status = ?", int32(req.Status))
	}
	if pageFilter != nil {
		if len(pageFilter.StrategyUIDs) > 0 {
			tx = tx.Where("strategy_uid IN ?", pageFilter.StrategyUIDs)
		}
		if len(pageFilter.LevelUIDs) > 0 {
			tx = tx.Where("level_uid IN ?", pageFilter.LevelUIDs)
		}
		if len(pageFilter.StrategyGroupUIDs) > 0 {
			tx = tx.Where("strategy_group_uid IN ?", pageFilter.StrategyGroupUIDs)
		}
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, err
	}
	req.WithTotal(total)
	tx = tx.Order("fired_at DESC")
	if req.Page > 0 && req.PageSize > 0 {
		tx = tx.Offset(req.Offset()).Limit(req.Limit())
	}
	var list []*do.AlertEvent
	if err := tx.Find(&list).Error; err != nil {
		return nil, err
	}

	levelNames := r.levelNamesForEvents(ctx, ns, list)
	items := make([]*bo.AlertEventItemBo, 0, len(list))
	for _, m := range list {
		name := levelNames[m.LevelUID.Int64()]
		items = append(items, convert.ToAlertEventItemBo(m, name))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func (r *alertEventRepository) levelNamesForEvents(ctx context.Context, namespaceUID snowflake.ID, events []*do.AlertEvent) map[int64]string {
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
	).Find()
	if err != nil {
		return nil
	}
	out := make(map[int64]string, len(levels))
	for _, lev := range levels {
		out[lev.ID.Int64()] = lev.Name
	}
	return out
}

func (r *alertEventRepository) InterveneAlert(ctx context.Context, uid snowflake.ID, by snowflake.ID) error {
	ns := contextx.GetNamespace(ctx)
	tableName := do.GenAlertEventTableName(ns, do.AlertEventTimeFromID(uid))
	if _, err := r.Cache().Get(ctx, cache.K(tableName)); err != nil && !r.DB().Migrator().HasTable(tableName) {
		return merr.ErrorNotFound("alert event not found")
	}
	bizQuery := query.Use(r.DB().Table(tableName))
	e := bizQuery.AlertEvent
	table := e.As(tableName)
	now := time.Now()
	info, err := e.WithContext(ctx).Where(table.ID.Eq(uid.Int64())).UpdateColumnSimple(
		e.Status.Value(do.AlertEventStatusIntervened),
		e.IntervenedAt.Value(now),
		e.IntervenedBy.Value(by.Int64()),
	)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("alert event not found")
	}
	return nil
}

func (r *alertEventRepository) SuppressAlert(ctx context.Context, uid snowflake.ID, until time.Time) error {
	ns := contextx.GetNamespace(ctx)
	tableName := do.GenAlertEventTableName(ns, do.AlertEventTimeFromID(uid))
	if _, err := r.Cache().Get(ctx, cache.K(tableName)); err != nil && !r.DB().Migrator().HasTable(tableName) {
		return merr.ErrorNotFound("alert event not found")
	}
	bizQuery := query.Use(r.DB().Table(tableName))
	e := bizQuery.AlertEvent
	table := e.As(tableName)
	info, err := e.WithContext(ctx).Where(table.ID.Eq(uid.Int64())).UpdateColumnSimple(
		e.Status.Value(do.AlertEventStatusSuppressed),
		e.SuppressedUntil.Value(until),
	)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("alert event not found")
	}
	return nil
}

func (r *alertEventRepository) RecoverAlert(ctx context.Context, uid snowflake.ID, by snowflake.ID) error {
	ns := contextx.GetNamespace(ctx)
	tableName := do.GenAlertEventTableName(ns, do.AlertEventTimeFromID(uid))
	if _, err := r.Cache().Get(ctx, cache.K(tableName)); err != nil && !r.DB().Migrator().HasTable(tableName) {
		return merr.ErrorNotFound("alert event not found")
	}
	bizQuery := query.Use(r.DB().Table(tableName))
	e := bizQuery.AlertEvent
	table := e.As(tableName)
	now := time.Now()
	info, err := e.WithContext(ctx).Where(table.ID.Eq(uid.Int64())).UpdateColumnSimple(
		e.Status.Value(do.AlertEventStatusRecovered),
		e.RecoveredAt.Value(now),
		e.RecoveredBy.Value(by.Int64()),
	)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("alert event not found")
	}
	return nil
}
