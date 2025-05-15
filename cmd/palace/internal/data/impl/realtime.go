package impl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/event"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/palace/internal/data"
	"github.com/moon-monitor/moon/cmd/palace/internal/helper/permission"
	"github.com/moon-monitor/moon/pkg/merr"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func NewRealtime(data *data.Data) repository.Realtime {
	return &realtimeImpl{
		Data: data,
	}
}

type realtimeImpl struct {
	*data.Data
}

func (r *realtimeImpl) getRealtimeTableName(ctx context.Context, alertStartsAt time.Time) (string, error) {
	teamId, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return "", merr.ErrorPermissionDenied("team id not found")
	}
	eventDB, err := r.GetEventDB(teamId)
	if err != nil {
		return "", err
	}
	tableName, err := event.GetRealtimeTableName(teamId, alertStartsAt, eventDB.GetDB())
	if err != nil {
		return "", err
	}
	return tableName, nil
}

// Exists implements repository.Realtime.
func (r *realtimeImpl) Exists(ctx context.Context, alert *bo.GetAlertParams) (bool, error) {
	ctx = permission.WithTeamIDContext(ctx, alert.TeamID)
	tx, teamId := getTeamEventQueryWithTeamID(ctx, r)
	tableName, err := r.getRealtimeTableName(ctx, alert.StartsAt)
	if err != nil {
		return false, err
	}
	realtimeQuery := tx.Realtime.Table(tableName)
	wrappers := []gen.Condition{
		realtimeQuery.Fingerprint.Eq(alert.Fingerprint),
		realtimeQuery.TeamID.Eq(teamId),
	}

	count, err := realtimeQuery.WithContext(ctx).
		Where(wrappers...).
		Limit(1).
		Count()
	if err != nil {
		return false, err
	}
	return count == 1, nil
}

// GetAlert implements repository.Realtime.
func (r *realtimeImpl) GetAlert(ctx context.Context, alert *bo.GetAlertParams) (do.Realtime, error) {
	ctx = permission.WithTeamIDContext(ctx, alert.TeamID)
	tx, teamId := getTeamEventQueryWithTeamID(ctx, r)
	tableName, err := r.getRealtimeTableName(ctx, alert.StartsAt)
	if err != nil {
		return nil, err
	}
	realtimeQuery := tx.Realtime.Table(tableName)
	wrappers := []gen.Condition{
		realtimeQuery.Fingerprint.Eq(alert.Fingerprint),
		realtimeQuery.TeamID.Eq(teamId),
	}

	realtimeDo, err := realtimeQuery.WithContext(ctx).
		Where(wrappers...).
		First()
	if err != nil {
		return nil, realtimeNotFound(err)
	}
	return realtimeDo, nil
}

// CreateAlert implements repository.Realtime.
func (r *realtimeImpl) CreateAlert(ctx context.Context, alert *bo.Alert) error {
	ctx = permission.WithTeamIDContext(ctx, alert.TeamID)
	tx, teamId := getTeamEventQueryWithTeamID(ctx, r)

	tableName, err := r.getRealtimeTableName(ctx, alert.StartsAt)
	if err != nil {
		return err
	}
	realtimeMutation := tx.Realtime.Table(tableName)
	realtimeDo := &event.Realtime{
		TeamID:       teamId,
		Fingerprint:  alert.Fingerprint,
		Labels:       alert.Labels,
		Summary:      alert.Summary,
		Description:  alert.Description,
		Value:        alert.Value,
		Status:       alert.Status,
		GeneratorURL: alert.GeneratorURL,
		StartsAt:     alert.StartsAt,
		EndsAt:       alert.EndsAt,
	}
	return realtimeMutation.WithContext(ctx).Create(realtimeDo)
}

// UpdateAlert implements repository.Realtime.
func (r *realtimeImpl) UpdateAlert(ctx context.Context, alert *bo.Alert) error {
	ctx = permission.WithTeamIDContext(ctx, alert.TeamID)
	tx, teamId := getTeamEventQueryWithTeamID(ctx, r)
	tableName, err := r.getRealtimeTableName(ctx, alert.StartsAt)
	if err != nil {
		return err
	}
	realtimeMutation := tx.Realtime.Table(tableName)
	wrappers := []gen.Condition{
		realtimeMutation.Fingerprint.Eq(alert.Fingerprint),
		realtimeMutation.TeamID.Eq(teamId),
	}
	mutations := []field.AssignExpr{
		realtimeMutation.Status.Value(alert.Status.GetValue()),
		realtimeMutation.GeneratorURL.Value(alert.GeneratorURL),
	}
	if alert.Status.IsResolved() {
		mutations = append(mutations, realtimeMutation.EndsAt.Value(alert.EndsAt))
	}
	_, err = realtimeMutation.WithContext(ctx).
		Where(wrappers...).
		UpdateSimple(mutations...)
	if err != nil {
		return err
	}
	return nil
}

// ListAlerts implements repository.Realtime.
func (r *realtimeImpl) ListAlerts(ctx context.Context, params *bo.ListAlertParams) (*bo.ListAlertReply, error) {
	bizDB, err := r.GetBizDB(params.TeamID)
	if err != nil {
		return nil, err
	}
	tableNames := event.GetRealtimeTableNames(params.TeamID, params.TimeRange[0], params.TimeRange[1], bizDB.GetDB())

	tables := make([]any, 0, len(tableNames))
	unionAllSQL := make([]string, 0, len(tableNames))
	for _, tableName := range tableNames {
		tables = append(tables, r.buildWrapper(bizDB.GetDB().Table(tableName), params))
		unionAllSQL = append(unionAllSQL, "?")
	}

	queryDB := bizDB.GetDB().Table(fmt.Sprintf("(%s) as combined_results", strings.Join(unionAllSQL, " UNION ALL ")), tables...)
	var realtimeDo []*event.Realtime
	queryDB = r.buildWrapper(queryDB, params)
	if validate.IsNotNil(params.PaginationRequest) {
		var total int64
		if err = queryDB.WithContext(ctx).Count(&total).Error; err != nil {
			return nil, err
		}
		params.WithTotal(total)
		queryDB = queryDB.Limit(int(params.Limit)).Offset(params.Offset())
	}
	err = queryDB.WithContext(ctx).Order("created_at DESC").Find(&realtimeDo).Error
	if err != nil {
		return nil, err
	}
	return params.ToListAlertReply(realtimeDo), nil
}

func (r *realtimeImpl) buildWrapper(bizDB *gorm.DB, params *bo.ListAlertParams) *gorm.DB {
	if params.Keyword != "" {
		bizDB = bizDB.Where("summary LIKE ? or description LIKE ?", params.Keyword, params.Keyword)
	}
	if params.Fingerprint != "" {
		bizDB = bizDB.Where("fingerprint = ?", params.Fingerprint)
	}
	if !params.Status.IsUnknown() {
		bizDB = bizDB.Where("status = ?", params.Status.GetValue())
	}
	bizDB = bizDB.Where("team_id = ?", params.TeamID)
	bizDB = bizDB.Where("starts_at >= ?", params.TimeRange[0])
	bizDB = bizDB.Where("starts_at <= ?", params.TimeRange[1])
	return bizDB
}
