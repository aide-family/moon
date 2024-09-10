package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// NewAlarmHistoryRepository 创建告警记录仓库
func NewAlarmHistoryRepository(data *data.Data) repository.HistoryRepository {
	return &alarmHistoryRepositoryImpl{
		data: data,
	}
}

type (
	alarmHistoryRepositoryImpl struct {
		data *data.Data
	}
)

func (a *alarmHistoryRepositoryImpl) GetAlarmHistory(ctx context.Context, param *bo.GetAlarmHistoryParams) (*alarmmodel.AlarmHistory, error) {
	alarmQuery, err := getBizAlarmQuery(ctx, a.data)
	if err != nil {
		return nil, err
	}
	alarmHistory, err := alarmQuery.AlarmHistory.WithContext(ctx).Where(alarmQuery.AlarmHistory.ID.Eq(param.ID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nAlarmHistoryDataNotFoundErr(ctx).WithCause(err)
		}
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return alarmHistory, nil
}

func (a *alarmHistoryRepositoryImpl) GetAlarmHistories(ctx context.Context, param *bo.QueryAlarmHistoryListParams) ([]*alarmmodel.AlarmHistory, error) {
	alarmQuery, err := getBizAlarmQuery(ctx, a.data)
	if err != nil {
		return nil, err
	}
	bizWrapper := alarmQuery.AlarmHistory.WithContext(ctx)
	var wheres []gen.Condition
	if !types.TextIsNull(param.InstanceName) {
		wheres = append(wheres, alarmQuery.AlarmHistory.InstanceName.Like(param.InstanceName))
	}

	if !param.AlertStatus.IsUnknown() {
		wheres = append(wheres, alarmQuery.AlarmHistory.Status.Like(param.AlertStatus.GetValue()))
	}

	if !types.TextIsNull(param.Keyword) {
		bizWrapper = bizWrapper.Or(alarmQuery.AlarmHistory.InstanceName.Like(param.Keyword))
		bizWrapper = bizWrapper.Or(alarmQuery.AlarmHistory.Info.Like(param.Keyword))
	}

	bizWrapper = bizWrapper.Where(wheres...)
	if err := types.WithPageQuery[bizquery.IAlarmHistoryDo](bizWrapper, param.Page); err != nil {
		return nil, err
	}
	return bizWrapper.Order(alarmQuery.AlarmHistory.ID.Desc()).Find()
}
