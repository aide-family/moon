package impl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gen"
	"gorm.io/gorm"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/event"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewSendMessageLog(data *data.Data) repository.SendMessageLog {
	return &sendMessageLogImpl{
		Data: data,
	}
}

type sendMessageLogImpl struct {
	*data.Data
}

func (s *sendMessageLogImpl) Retry(ctx context.Context, params *bo.RetrySendMessageParams) error {
	if params.TeamID > 0 {
		return s.retryTeamSendMessageLog(ctx, params)
	}
	return s.retrySystemSendMessageLog(ctx, params)
}

func (s *sendMessageLogImpl) getTeamSendMessageLogTableName(ctx context.Context, sendAt time.Time) (string, error) {
	teamId, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return "", merr.ErrorPermissionDenied("team id not found")
	}
	bizDB, err := s.GetEventDB(teamId)
	if err != nil {
		return "", err
	}
	return event.GetSendMessageLogTableName(teamId, sendAt, bizDB.GetDB())
}

func (s *sendMessageLogImpl) getSystemSendMessageLogTableName(sendAt time.Time) (string, error) {
	tx := s.GetMainDB()
	return system.GetSendMessageLogTableName(sendAt, tx.GetDB())
}

func (s *sendMessageLogImpl) retryTeamSendMessageLog(ctx context.Context, params *bo.RetrySendMessageParams) error {
	tx, teamId := getTeamEventQueryWithTeamID(ctx, s)
	tableName, err := s.getTeamSendMessageLogTableName(ctx, params.SendAt)
	if err != nil {
		return err
	}
	sendMessageLogTx := tx.SendMessageLog.Table(tableName)
	wrapper := sendMessageLogTx.WithContext(ctx)
	wrapper = wrapper.Where(sendMessageLogTx.TeamID.Eq(teamId), sendMessageLogTx.RequestID.Eq(params.RequestID))
	_, err = wrapper.UpdateSimple(sendMessageLogTx.RetryCount.Add(1))
	return err
}

func (s *sendMessageLogImpl) retrySystemSendMessageLog(ctx context.Context, params *bo.RetrySendMessageParams) error {
	tx := getMainQuery(ctx, s)
	tableName, err := s.getSystemSendMessageLogTableName(params.SendAt)
	if err != nil {
		return err
	}
	sendMessageLogTx := tx.SendMessageLog.Table(tableName)
	wrapper := sendMessageLogTx.WithContext(ctx)
	wrapper = wrapper.Where(sendMessageLogTx.RequestID.Eq(params.RequestID))
	_, err = wrapper.UpdateSimple(sendMessageLogTx.RetryCount.Add(1))
	return err
}

// List implements repository.SendMessageLog.
func (s *sendMessageLogImpl) List(ctx context.Context, params *bo.ListSendMessageLogParams) (*bo.ListSendMessageLogReply, error) {
	if params.TeamID > 0 {
		return s.listTeamSendMessageLog(ctx, params)
	}
	return s.listSystemSendMessageLog(ctx, params)
}

// Get implements repository.SendMessageLog.
func (s *sendMessageLogImpl) Get(ctx context.Context, params *bo.GetSendMessageLogParams) (do.SendMessageLog, error) {
	if params.TeamID > 0 {
		return s.getTeamSendMessageLog(ctx, params)
	}
	return s.getSystemSendMessageLog(ctx, params)
}

// UpdateStatus implements repository.SendMessageLog.
func (s *sendMessageLogImpl) UpdateStatus(ctx context.Context, params *bo.UpdateSendMessageLogStatusParams) error {
	if params.TeamID > 0 {
		return s.updateTeamSendMessageLog(ctx, params)
	}
	return s.updateSystemSendMessageLog(ctx, params)
}

func (s *sendMessageLogImpl) Create(ctx context.Context, params *bo.CreateSendMessageLogParams) error {
	if params.TeamID > 0 {
		return s.createTeamSendMessageLog(ctx, params)
	}
	return s.createSystemSendMessageLog(ctx, params)
}

func (s *sendMessageLogImpl) createTeamSendMessageLog(ctx context.Context, params *bo.CreateSendMessageLogParams) error {
	sendMessageLog := &event.SendMessageLog{
		TeamID:      params.TeamID,
		MessageType: params.MessageType,
		Message:     params.Message.String(),
		RequestID:   params.RequestID,
		Status:      vobj.SendMessageStatusSending,
		RetryCount:  0,
		Error:       "",
		SentAt:      params.SendAt,
	}
	sendMessageLog.WithContext(ctx)
	tx := getTeamEventQuery(ctx, s)
	tableName, err := s.getTeamSendMessageLogTableName(ctx, params.SendAt)
	if err != nil {
		return err
	}
	sendMessageLogTx := tx.SendMessageLog.Table(tableName)
	return sendMessageLogTx.WithContext(ctx).Create(sendMessageLog)
}

func (s *sendMessageLogImpl) createSystemSendMessageLog(ctx context.Context, params *bo.CreateSendMessageLogParams) error {
	sendMessageLog := &system.SendMessageLog{
		SentAt:      params.SendAt,
		MessageType: params.MessageType,
		Message:     params.Message.String(),
		RequestID:   params.RequestID,
		Status:      vobj.SendMessageStatusSending,
		RetryCount:  0,
		Error:       "",
	}
	sendMessageLog.WithContext(ctx)
	tx := getMainQuery(ctx, s)
	tableName, err := s.getSystemSendMessageLogTableName(params.SendAt)
	if err != nil {
		return err
	}
	sendMessageLogTx := tx.SendMessageLog.Table(tableName)
	return sendMessageLogTx.WithContext(ctx).Create(sendMessageLog)
}

func (s *sendMessageLogImpl) getTeamSendMessageLog(ctx context.Context, params *bo.GetSendMessageLogParams) (do.SendMessageLog, error) {
	ctx = permission.WithTeamIDContext(ctx, params.TeamID)
	tx, teamId := getTeamEventQueryWithTeamID(ctx, s)
	tableName, err := s.getTeamSendMessageLogTableName(ctx, params.SendAt)
	if err != nil {
		return nil, err
	}
	sendMessageLogTx := tx.SendMessageLog.Table(tableName)
	wrapper := sendMessageLogTx.WithContext(ctx)
	wrappers := []gen.Condition{
		sendMessageLogTx.TeamID.Eq(teamId),
		sendMessageLogTx.RequestID.Eq(params.RequestID),
	}
	sendMessageLog, err := wrapper.Where(wrappers...).First()
	if err != nil {
		return nil, sendMessageLogNotFound(err)
	}
	return sendMessageLog, nil
}

func (s *sendMessageLogImpl) getSystemSendMessageLog(ctx context.Context, params *bo.GetSendMessageLogParams) (do.SendMessageLog, error) {
	tx := getMainQuery(ctx, s)
	tableName, err := s.getSystemSendMessageLogTableName(params.SendAt)
	if err != nil {
		return nil, err
	}
	sendMessageLogTx := tx.SendMessageLog.Table(tableName)
	wrapper := sendMessageLogTx.WithContext(ctx)
	wrappers := []gen.Condition{
		sendMessageLogTx.RequestID.Eq(params.RequestID),
	}
	sendMessageLog, err := wrapper.Where(wrappers...).First()
	if err != nil {
		return nil, sendMessageLogNotFound(err)
	}
	return sendMessageLog, nil
}

func (s *sendMessageLogImpl) updateTeamSendMessageLog(ctx context.Context, params *bo.UpdateSendMessageLogStatusParams) error {
	ctx = permission.WithTeamIDContext(ctx, params.TeamID)
	tx, teamId := getTeamEventQueryWithTeamID(ctx, s)
	tableName, err := s.getTeamSendMessageLogTableName(ctx, params.SendAt)
	if err != nil {
		return err
	}
	sendMessageLogTx := tx.SendMessageLog.Table(tableName)
	wrapper := sendMessageLogTx.WithContext(ctx)
	wrappers := []gen.Condition{
		sendMessageLogTx.TeamID.Eq(teamId),
		sendMessageLogTx.RequestID.Eq(params.RequestID),
	}
	sendMessageLog, err := wrapper.Where(wrappers...).First()
	if err != nil {
		return sendMessageLogNotFound(err)
	}
	sendMessageLog.WithContext(ctx)
	sendMessageLog.Status = params.Status
	sendMessageLog.Error = strings.Join([]string{sendMessageLog.Error, params.Error}, "\n")
	return wrapper.Save(sendMessageLog)
}

func (s *sendMessageLogImpl) updateSystemSendMessageLog(ctx context.Context, params *bo.UpdateSendMessageLogStatusParams) error {
	tx := getMainQuery(ctx, s)
	tableName, err := s.getSystemSendMessageLogTableName(params.SendAt)
	if err != nil {
		return err
	}
	sendMessageLogTx := tx.SendMessageLog.Table(tableName)
	wrapper := sendMessageLogTx.WithContext(ctx)
	wrappers := []gen.Condition{
		sendMessageLogTx.RequestID.Eq(params.RequestID),
	}
	sendMessageLog, err := wrapper.Where(wrappers...).First()
	if err != nil {
		return sendMessageLogNotFound(err)
	}
	sendMessageLog.WithContext(ctx)
	sendMessageLog.Status = params.Status
	sendMessageLog.Error = strings.Join([]string{sendMessageLog.Error, params.Error}, "\n")
	return wrapper.Save(sendMessageLog)
}

func (s *sendMessageLogImpl) listTeamSendMessageLog(ctx context.Context, params *bo.ListSendMessageLogParams) (*bo.ListSendMessageLogReply, error) {
	eventDB, err := s.GetEventDB(params.TeamID)
	if err != nil {
		return nil, err
	}
	startAt, endAt := params.TimeRange[0], params.TimeRange[1]
	if startAt.IsZero() {
		startAt = timex.Now().AddDate(0, 0, -7)
	}
	if endAt.IsZero() {
		endAt = timex.Now()
	}
	tableNames := event.GetSendMessageLogTableNames(params.TeamID, startAt, endAt, eventDB.GetDB())
	tables := make([]any, 0, len(tableNames))
	unionAllSQL := make([]string, 0, len(tableNames))
	for _, tableName := range tableNames {
		tables = append(tables, eventDB.GetDB().Table(tableName))
		unionAllSQL = append(unionAllSQL, "?")
	}
	var sendMessageLogs []*event.SendMessageLog
	queryDB := eventDB.GetDB().Table(fmt.Sprintf("(%s) as combined_results", strings.Join(unionAllSQL, " UNION ALL ")), tables...)
	queryDB = s.buildSendMessageLogWrapper(queryDB, params)
	queryDB = queryDB.Where("team_id = ?", params.TeamID)
	if validate.IsNotNil(params.PaginationRequest) {
		var total int64
		if err = queryDB.WithContext(ctx).Count(&total).Error; err != nil {
			return nil, err
		}
		params.WithTotal(total)
		queryDB = queryDB.Limit(int(params.Limit)).Offset(params.Offset())
	}
	err = queryDB.WithContext(ctx).Order("created_at DESC").Find(&sendMessageLogs).Error
	if err != nil {
		return nil, err
	}
	rows := slices.Map(sendMessageLogs, func(log *event.SendMessageLog) do.SendMessageLog {
		return log
	})
	return params.ToListReply(rows), nil
}

func (s *sendMessageLogImpl) buildSendMessageLogWrapper(eventDB *gorm.DB, params *bo.ListSendMessageLogParams) *gorm.DB {
	if validate.TextIsNotNull(params.Keyword) {
		eventDB = eventDB.Where("message LIKE ?", params.Keyword)
	}
	if validate.TextIsNotNull(params.RequestID) {
		eventDB = eventDB.Where("request_id = ?", params.RequestID)
	}
	if !params.MessageType.IsUnknown() {
		eventDB = eventDB.Where("message_type = ?", params.MessageType.GetValue())
	}
	if len(params.TimeRange) == 2 {
		eventDB = eventDB.Where("created_at BETWEEN ? AND ?", params.TimeRange[0], params.TimeRange[1])
	}
	if !params.Status.IsUnknown() {
		eventDB = eventDB.Where("status = ?", params.Status.GetValue())
	}
	return eventDB
}

func (s *sendMessageLogImpl) listSystemSendMessageLog(ctx context.Context, params *bo.ListSendMessageLogParams) (*bo.ListSendMessageLogReply, error) {
	mainDB := s.GetMainDB().GetDB()
	startAt, endAt := params.TimeRange[0], params.TimeRange[1]
	if startAt.IsZero() {
		startAt = timex.Now().AddDate(0, 0, -7)
	}
	if endAt.IsZero() {
		endAt = timex.Now()
	}
	tableNames := system.GetSendMessageLogTableNames(startAt, endAt, mainDB)
	tables := make([]any, 0, len(tableNames))
	unionAllSQL := make([]string, 0, len(tableNames))
	for _, tableName := range tableNames {
		tables = append(tables, mainDB.Table(tableName))
		unionAllSQL = append(unionAllSQL, "?")
	}
	var sendMessageLogs []*system.SendMessageLog
	queryDB := mainDB.Table(fmt.Sprintf("(%s) as combined_results", strings.Join(unionAllSQL, " UNION ALL ")), tables...)
	queryDB = s.buildSendMessageLogWrapper(queryDB, params)
	if validate.IsNotNil(params.PaginationRequest) {
		var total int64
		if err := queryDB.WithContext(ctx).Count(&total).Error; err != nil {
			return nil, err
		}
		params.WithTotal(total)
		queryDB = queryDB.Limit(int(params.Limit)).Offset(params.Offset())
	}
	err := queryDB.WithContext(ctx).Order("created_at DESC").Find(&sendMessageLogs).Error
	if err != nil {
		return nil, err
	}
	rows := slices.Map(sendMessageLogs, func(log *system.SendMessageLog) do.SendMessageLog {
		return log
	})
	return params.ToListReply(rows), nil
}
