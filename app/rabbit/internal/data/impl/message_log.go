package impl

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/plugin/cache"
	"github.com/aide-family/magicbox/pointer"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/data"
	"github.com/aide-family/rabbit/internal/data/impl/convert"
	"github.com/aide-family/rabbit/internal/data/impl/do"
	"github.com/aide-family/rabbit/internal/data/impl/query"
)

// messageLogTableCreateMu serializes creation of message_logs base table and rename,
// to avoid "index already exists" when GORM CreateTable runs concurrently (SQLite does not use CREATE INDEX IF NOT EXISTS).
var messageLogTableCreateMu sync.Mutex

func NewMessageLogRepository(d *data.Data) repository.MessageLog {
	query.SetDefault(d.DB())
	return &messageLogRepository{Data: d}
}

type messageLogRepository struct {
	*data.Data
}

// GetMessageLog implements [repository.MessageLog].
func (m *messageLogRepository) GetMessageLog(ctx context.Context, uid snowflake.ID) (*bo.MessageLogItemBo, error) {
	return m.getMessageLog(ctx, uid)
}

// GetAllMessageLogs implements [repository.MessageLog].
func (m *messageLogRepository) GetAllMessageLogs(ctx context.Context, status enum.MessageStatus) ([]*bo.MessageLogItemBo, error) {
	namespace := contextx.GetNamespace(ctx)
	tableName := do.GenMessageLogTableName(namespace, time.Now())
	if _, err := m.Cache().Get(ctx, cache.K(tableName)); err != nil && !do.HasTable(m.DB(), tableName) {
		return []*bo.MessageLogItemBo{}, nil
	}

	bizQuery := query.Use(m.DB().Table(tableName))
	messageLog := bizQuery.MessageLog
	messageLogTable := messageLog.As(tableName)
	wrappers := messageLog.WithContext(ctx)
	wheres := []gen.Condition{
		messageLogTable.Status.Eq(int32(status)),
	}
	wrappers = wrappers.Where(wheres...)
	messageLogs, err := wrappers.Order(messageLogTable.CreatedAt.Asc()).Find()
	if err != nil {
		return nil, err
	}
	messageLogItems := make([]*bo.MessageLogItemBo, 0, len(messageLogs))
	for _, messageLog := range messageLogs {
		messageLogItems = append(messageLogItems, convert.ToMessageLogItemBo(messageLog))
	}
	return messageLogItems, nil
}

// GetMessageLogWithLock implements [repository.MessageLog].
func (m *messageLogRepository) GetMessageLogWithLock(ctx context.Context, uid snowflake.ID) (*bo.MessageLogItemBo, error) {
	return m.getMessageLog(ctx, uid, clause.Locking{Strength: "UPDATE"})
}

func (m *messageLogRepository) getMessageLog(ctx context.Context, uid snowflake.ID, clauses ...clause.Expression) (*bo.MessageLogItemBo, error) {
	namespace := contextx.GetNamespace(ctx)
	tableName := do.GenMessageLogTableName(namespace, time.UnixMilli(uid.Time()))
	if _, err := m.Cache().Get(ctx, cache.K(tableName)); err != nil && !do.HasTable(m.DB(), tableName) {
		return nil, gorm.ErrRecordNotFound
	}

	bizQuery := query.Use(m.DB().Table(tableName))
	messageLog := bizQuery.MessageLog
	messageLogTable := messageLog.As(tableName)
	wrappers := messageLog.WithContext(ctx)
	wheres := []gen.Condition{
		messageLogTable.ID.Eq(uid.Int64()),
		messageLogTable.NamespaceUID.Eq(namespace.Int64()),
	}
	wrappers = wrappers.Where(wheres...).Clauses(clauses...)
	messageLogDO, err := wrappers.First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("message log %d not found", uid.Int64())
		}
		return nil, err
	}
	return convert.ToMessageLogItemBo(messageLogDO), nil
}

// ListMessageLog implements [repository.MessageLog].
func (m *messageLogRepository) ListMessageLog(ctx context.Context, req *bo.ListMessageLogBo) (*bo.PageResponseBo[*bo.MessageLogItemBo], error) {
	namespace := contextx.GetNamespace(ctx)

	if req.StartAt.IsZero() {
		req.StartAt = time.Now().AddDate(0, 0, -7)
	}
	if req.EndAt.IsZero() {
		req.EndAt = time.Now()
	}
	bizDB := m.DB()

	tableNames := do.GenMessageLogTableNames(bizDB, namespace, req.StartAt, req.EndAt)
	if len(tableNames) == 0 {
		return bo.NewPageResponseBo[*bo.MessageLogItemBo](req.PageRequestBo, nil), nil
	}

	tables := make([]any, 0, len(tableNames))
	unionAllSQL := make([]string, 0, len(tableNames))
	for _, tableName := range tableNames {
		tables = append(tables, bizDB.Table(tableName))
		unionAllSQL = append(unionAllSQL, "?")
	}
	wrappers := bizDB.WithContext(ctx)
	if len(tableNames) > 1 {
		wrappers = wrappers.Table(fmt.Sprintf("(%s) as %s", strings.Join(unionAllSQL, " UNION ALL "), do.TableNameMessageLog), tables...)
	} else {
		wrappers = wrappers.Table(fmt.Sprintf("%s as %s", tableNames[0], do.TableNameMessageLog))
	}

	messageLog := query.MessageLog
	messageLogTable := messageLog.As(do.TableNameMessageLog)

	wrappers = wrappers.Where(messageLogTable.SendAt.Gte(req.StartAt))
	wrappers = wrappers.Where(messageLogTable.SendAt.Lte(req.EndAt))
	wrappers = wrappers.Where(messageLogTable.NamespaceUID.Eq(namespace.Int64()))

	if req.Status > enum.MessageStatus_MessageStatus_UNKNOWN {
		wrappers = wrappers.Where(messageLogTable.Status.Eq(int32(req.Status)))
	}
	if req.MessageType > enum.MessageType_MessageType_UNKNOWN {
		wrappers = wrappers.Where(messageLogTable.Type.Eq(int32(req.MessageType)))
	}
	if pointer.IsNotNil(req.PageRequestBo) {
		var total int64
		if err := wrappers.Count(&total).Error; err != nil {
			return nil, err
		}
		req.WithTotal(total)
		wrappers = wrappers.Limit(req.Limit()).Offset(req.Offset())
	}
	wrappers = wrappers.Order(clause.OrderByColumn{Column: clause.Column{Name: messageLogTable.SendAt.ColumnName().String()}, Desc: true})
	var messageLogs []*do.MessageLog
	if err := wrappers.Find(&messageLogs).Error; err != nil {
		return nil, err
	}
	messageLogItems := make([]*bo.MessageLogItemBo, 0, len(messageLogs))
	for _, messageLog := range messageLogs {
		messageLogItems = append(messageLogItems, convert.ToMessageLogItemBo(messageLog))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, messageLogItems), nil
}

// UpdateMessageLogStatusIf implements [repository.MessageLog].
func (m *messageLogRepository) UpdateMessageLogStatusIf(ctx context.Context, uid snowflake.ID, oldStatus enum.MessageStatus, newStatus enum.MessageStatus) (bool, error) {
	namespace := contextx.GetNamespace(ctx)
	tableName := do.GenMessageLogTableName(namespace, time.UnixMilli(uid.Time()))
	if _, err := m.Cache().Get(ctx, cache.K(tableName)); err != nil && !do.HasTable(m.DB(), tableName) {
		return false, merr.ErrorNotFound("message log %d not found", uid.Int64())
	}

	bizQuery := query.Use(m.DB().Table(tableName))
	messageLog := bizQuery.MessageLog
	messageLogTable := messageLog.As(tableName)
	wrappers := messageLog.WithContext(ctx)
	wheres := []gen.Condition{
		messageLogTable.ID.Eq(uid.Int64()),
		messageLogTable.NamespaceUID.Eq(namespace.Int64()),
		messageLogTable.Status.Eq(int32(oldStatus)),
	}
	wrappers = wrappers.Where(wheres...)
	result, err := wrappers.UpdateColumnSimple(messageLogTable.Status.Value(int32(newStatus)))
	if err != nil {
		return false, err
	}
	return result.RowsAffected == 1, nil
}

// UpdateMessageLogLastErrorIf implements [repository.MessageLog].
func (m *messageLogRepository) UpdateMessageLogLastErrorIf(ctx context.Context, uid snowflake.ID, oldStatus enum.MessageStatus, lastError string) (bool, error) {
	namespace := contextx.GetNamespace(ctx)
	tableName := do.GenMessageLogTableName(namespace, time.UnixMilli(uid.Time()))
	if _, err := m.Cache().Get(ctx, cache.K(tableName)); err != nil && !do.HasTable(m.DB(), tableName) {
		return false, merr.ErrorNotFound("message log %d not found", uid.Int64())
	}

	bizQuery := query.Use(m.DB().Table(tableName))
	messageLog := bizQuery.MessageLog
	messageLogTable := messageLog.As(tableName)
	wrappers := messageLog.WithContext(ctx)
	wheres := []gen.Condition{
		messageLogTable.ID.Eq(uid.Int64()),
		messageLogTable.NamespaceUID.Eq(namespace.Int64()),
		messageLogTable.Status.Eq(int32(oldStatus)),
	}
	wrappers = wrappers.Where(wheres...)
	columns := []field.AssignExpr{
		messageLogTable.LastError.Value(lastError),
		messageLogTable.Status.Value(int32(enum.MessageStatus_FAILED)),
	}
	result, err := wrappers.UpdateColumnSimple(columns...)
	if err != nil {
		return false, err
	}
	return result.RowsAffected == 1, nil
}

func (m *messageLogRepository) UpdateMessageLogStatusSuccessIf(ctx context.Context, uid snowflake.ID) (bool, error) {
	namespace := contextx.GetNamespace(ctx)
	tableName := do.GenMessageLogTableName(namespace, time.UnixMilli(uid.Time()))
	if _, err := m.Cache().Get(ctx, cache.K(tableName)); err != nil && !do.HasTable(m.DB(), tableName) {
		return false, merr.ErrorNotFound("message log %d not found", uid.Int64())
	}
	bizQuery := query.Use(m.DB().Table(tableName))
	messageLog := bizQuery.MessageLog
	messageLogTable := messageLog.As(tableName)
	wrappers := messageLog.WithContext(ctx)
	wheres := []gen.Condition{
		messageLogTable.ID.Eq(uid.Int64()),
		messageLogTable.NamespaceUID.Eq(namespace.Int64()),
	}
	wrappers = wrappers.Where(wheres...)
	result, err := wrappers.UpdateColumnSimple(messageLogTable.Status.Value(int32(enum.MessageStatus_SENT)))
	if err != nil {
		return false, err
	}
	return result.RowsAffected == 1, nil
}

func (m *messageLogRepository) UpdateMessageLogStatusSendingIf(ctx context.Context, uid snowflake.ID, oldStatus enum.MessageStatus) (bool, error) {
	namespace := contextx.GetNamespace(ctx)
	tableName := do.GenMessageLogTableName(namespace, time.UnixMilli(uid.Time()))
	if _, err := m.Cache().Get(ctx, cache.K(tableName)); err != nil && !do.HasTable(m.DB(), tableName) {
		return false, merr.ErrorNotFound("message log %d not found", uid.Int64())
	}
	bizQuery := query.Use(m.DB().Table(tableName))
	messageLog := bizQuery.MessageLog
	messageLogTable := messageLog.As(tableName)
	wrappers := messageLog.WithContext(ctx)
	wheres := []gen.Condition{
		messageLogTable.ID.Eq(uid.Int64()),
		messageLogTable.NamespaceUID.Eq(namespace.Int64()),
		messageLogTable.Status.Eq(int32(oldStatus)),
	}
	wrappers = wrappers.Where(wheres...)
	columns := []field.AssignExpr{
		messageLogTable.Status.Value(int32(enum.MessageStatus_SENDING)),
		messageLogTable.SendAt.Value(time.Now()),
	}
	result, err := wrappers.UpdateColumnSimple(columns...)
	if err != nil {
		return false, err
	}
	return result.RowsAffected == 1, nil
}

// CreateMessageLog implements [repository.MessageLog].
func (m *messageLogRepository) CreateMessageLog(ctx context.Context, req *bo.CreateMessageLogBo) (snowflake.ID, error) {
	messageLogDo := convert.ToMessageLogDo(ctx, req)
	tableName, err := m.getTableName(ctx, time.Now())
	if err != nil {
		return 0, err
	}
	bizQuery := query.Use(m.DB().Table(tableName))
	messageLog := bizQuery.MessageLog
	mutation := messageLog.WithContext(ctx)
	if err := mutation.Create(messageLogDo); err != nil {
		return 0, err
	}
	return messageLogDo.ID, nil
}

func (m *messageLogRepository) getTableName(ctx context.Context, timeAt time.Time) (string, error) {
	namespace := contextx.GetNamespace(ctx)
	tableName := do.GenMessageLogTableName(namespace, timeAt)

	if _, err := m.Cache().Get(ctx, cache.K(tableName)); err == nil && do.HasTable(m.DB(), tableName) {
		return tableName, nil
	}
	if !do.HasTable(m.DB(), tableName) {
		messageLogTableCreateMu.Lock()
		defer messageLogTableCreateMu.Unlock()
		initModel := &do.MessageLog{}
		oldTableName := initModel.TableName()
		if !do.HasTable(m.DB(), oldTableName) {
			if err := m.DB().Migrator().CreateTable(initModel); err != nil {
				return "", err
			}
		}
		if err := m.DB().Migrator().RenameTable(oldTableName, tableName); err != nil {
			return "", err
		}
	}
	if err := m.Cache().Set(ctx, cache.K(tableName), "", 0); err != nil {
		klog.Context(ctx).Warnw("msg", "set cache failed", "error", err, "tableName", tableName)
	}

	return tableName, nil
}

func (m *messageLogRepository) MessageLogRetryIncrement(ctx context.Context, uid snowflake.ID) error {
	namespace := contextx.GetNamespace(ctx)
	tableName := do.GenMessageLogTableName(namespace, time.UnixMilli(uid.Time()))
	if _, err := m.Cache().Get(ctx, cache.K(tableName)); err != nil && !do.HasTable(m.DB(), tableName) {
		return merr.ErrorNotFound("message log %d not found", uid.Int64())
	}
	bizQuery := query.Use(m.DB().Table(tableName))
	messageLog := bizQuery.MessageLog
	messageLogTable := messageLog.As(tableName)
	wrappers := messageLog.WithContext(ctx)
	_, err := wrappers.UpdateColumnSimple(messageLogTable.RetryTotal.Add(1), messageLogTable.Status.Value(int32(enum.MessageStatus_PENDING)))
	return err
}
