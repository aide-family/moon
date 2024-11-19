package repoimpl

import (
	"context"
	"time"

	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

// NewAlarmSendRepository 创建告警发送记录
func NewAlarmSendRepository(data *data.Data) repository.AlarmSendRepository {
	return &alarmSendRepositoryImpl{
		data: data,
	}
}

type (
	alarmSendRepositoryImpl struct {
		data *data.Data
	}
)

func (a *alarmSendRepositoryImpl) GetRetryNumberByRequestID(ctx context.Context, requestID string, teamID uint32) (int, error) {
	alarmQuery, err := getTeamBizAlarmQuery(teamID, a.data)
	if err != nil {
		return 0, err
	}
	history, err := alarmQuery.AlarmSendHistory.WithContext(ctx).Where(alarmQuery.AlarmSendHistory.RequestID.Eq(requestID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, merr.ErrorI18nToastAlarmSendHistoryNotFound(ctx).WithCause(err)
		}
		return 0, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return history.RetryNumber, nil
}

func (a *alarmSendRepositoryImpl) GetAlarmSendHistory(ctx context.Context, param *bo.GetAlarmSendHistoryParams) (*alarmmodel.AlarmSendHistory, error) {
	alarmQuery, err := getBizAlarmQuery(ctx, a.data)
	if err != nil {
		return nil, err
	}
	history, err := alarmQuery.AlarmSendHistory.WithContext(ctx).
		Preload(field.Associations).Where(alarmQuery.AlarmSendHistory.ID.Eq(param.ID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastAlarmSendHistoryNotFound(ctx).WithCause(err)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return history, nil
}

func (a *alarmSendRepositoryImpl) AlarmSendHistoryList(ctx context.Context, param *bo.QueryAlarmSendHistoryListParams) ([]*alarmmodel.AlarmSendHistory, error) {
	alarmQuery, err := getBizAlarmQuery(ctx, a.data)
	if err != nil {
		return nil, err
	}
	bizWrapper := alarmQuery.AlarmSendHistory.WithContext(ctx)
	var wheres []gen.Condition

	if !types.TextIsNull(param.StartSendTime) && !types.TextIsNull(param.EndSendTime) {
		wheres = append(wheres, alarmQuery.AlarmSendHistory.SendTime.Between(param.StartSendTime, param.EndSendTime))
	}

	if len(param.SendStatus) > 0 {
		sendStatus := types.SliceTo(param.SendStatus, func(status vobj.SendStatus) int {
			return status.GetValue()
		})
		wheres = append(wheres, alarmQuery.AlarmSendHistory.SendStatus.In(sendStatus...))
	}

	if !types.TextIsNull(param.Keyword) {
		bizWrapper = bizWrapper.Or(alarmQuery.AlarmSendHistory.SendData.Like(param.Keyword))
	}

	bizWrapper = bizWrapper.Where(wheres...).Order(alarmQuery.AlarmSendHistory.ID.Desc())

	if bizWrapper, err = types.WithPageQuery(bizWrapper, param.Page); err != nil {
		return nil, err
	}
	return bizWrapper.Find()
}

func (a *alarmSendRepositoryImpl) SaveAlarmSendHistory(ctx context.Context, param *bo.CreateAlarmSendParams) error {
	alarmQuery, err := getTeamBizAlarmQuery(param.TeamID, a.data)
	if err != nil {
		return err
	}
	requestID := param.RequestID

	if param.SendTime == nil {
		param.SendTime = types.NewTime(time.Now())
	}

	// 查询是否已经存在
	sendHistory, _ := alarmQuery.AlarmSendHistory.WithContext(ctx).Where(alarmQuery.AlarmSendHistory.RequestID.Eq(requestID)).First()
	if types.IsNil(sendHistory) {
		sendHistoryModel := createAlarmSendHistoryToModel(param)
		if err = alarmQuery.AlarmSendHistory.Create(sendHistoryModel); !types.IsNil(err) {
			return err
		}
	} else {
		sendHistory.RetryNumber++
		sendHistory.SendStatus = param.SendStatus
		sendHistory.SendTime = param.SendTime.GoString()
		if _, err := alarmQuery.AlarmSendHistory.WithContext(ctx).Updates(sendHistory); err != nil {
			return err
		}
	}
	return nil
}

func (a *alarmSendRepositoryImpl) RetryAlarmSend(ctx context.Context, param *bo.RetryAlarmSendParams) error {
	alarmQuery, err := getBizAlarmQuery(ctx, a.data)
	if err != nil {
		return err
	}
	history, err := alarmQuery.AlarmSendHistory.WithContext(ctx).Where(alarmQuery.AlarmSendHistory.RequestID.Eq(param.RequestID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRegistered) {
			return merr.ErrorI18nToastHistoryAlarmNotFound(ctx).WithCause(err)
		}
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	sendMsgRequest := &hookapi.SendMsgRequest{
		Json:      history.SendData,
		RequestID: history.RequestID,
		Route:     history.Route,
	}
	sendMsg := &bo.SendMsg{
		SendMsgRequest: sendMsgRequest,
	}
	a.data.GetAlertPersistenceDBQueue().Push(watch.NewMessage(sendMsg, vobj.TopicAlertMsg))
	return nil
}

func createAlarmSendHistoryToModel(param *bo.CreateAlarmSendParams) *alarmmodel.AlarmSendHistory {
	sendHistoryModel := &alarmmodel.AlarmSendHistory{
		SendStatus:   param.SendStatus,
		SendData:     param.SendData,
		RequestID:    param.RequestID,
		RetryNumber:  param.RetryNumber,
		AlarmGroupID: param.AlarmGroupID,
		SendTime:     param.SendTime.String(),
		Route:        param.Route,
	}
	return sendHistoryModel
}
