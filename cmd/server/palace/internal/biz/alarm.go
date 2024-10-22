package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/microrepository"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// NewAlarmBiz 创建告警相关业务逻辑
func NewAlarmBiz(
	alarmRepository repository.Alarm,
	alarmRawRepository repository.AlarmRaw,
	strategyRepository repository.Strategy,
	datasourceRepository repository.Datasource,
	historyRepository repository.HistoryRepository,
	sendAlert microrepository.SendAlert,
) *AlarmBiz {
	return &AlarmBiz{
		alarmRepository:      alarmRepository,
		alarmRawRepository:   alarmRawRepository,
		historyRepository:    historyRepository,
		strategyRepository:   strategyRepository,
		datasourceRepository: datasourceRepository,
		sendAlert:            sendAlert,
	}
}

// AlarmBiz 告警相关业务逻辑
type AlarmBiz struct {
	alarmRepository      repository.Alarm
	alarmRawRepository   repository.AlarmRaw
	historyRepository    repository.HistoryRepository
	strategyRepository   repository.Strategy
	datasourceRepository repository.Datasource
	sendAlert            microrepository.SendAlert
}

// GetRealTimeAlarm 获取实时告警明细
func (b *AlarmBiz) GetRealTimeAlarm(ctx context.Context, params *bo.GetRealTimeAlarmParams) (*alarmmodel.RealtimeAlarm, error) {
	return b.alarmRepository.GetRealTimeAlarm(ctx, params)
}

// ListRealTimeAlarms 获取实时告警列表
func (b *AlarmBiz) ListRealTimeAlarms(ctx context.Context, params *bo.GetRealTimeAlarmsParams) ([]*alarmmodel.RealtimeAlarm, error) {
	return b.alarmRepository.GetRealTimeAlarms(ctx, params)
}

// SaveAlertQueue 保存告警队列
func (b *AlarmBiz) SaveAlertQueue(param *bo.CreateAlarmHookRawParams) error {
	return b.alarmRepository.SaveAlertQueue(param)
}

// CreateAlarmRawInfo 创建告警原始信息
func (b *AlarmBiz) CreateAlarmRawInfo(ctx context.Context, param *bo.CreateAlarmHookRawParams) ([]*alarmmodel.AlarmRaw, error) {
	rawParamList := types.SliceTo(param.Alerts, func(item *bo.AlertItemRawParams) *bo.CreateAlarmRawParams {
		return &bo.CreateAlarmRawParams{
			Fingerprint: item.Fingerprint,
			RawInfo:     item.GetAlertItemString(),
			Receiver:    param.Receiver,
		}
	})
	return b.alarmRawRepository.CreateAlarmRaws(ctx, rawParamList, param.TeamID)
}

func (b *AlarmBiz) CreateAlarmInfo(ctx context.Context, params *bo.CreateAlarmHookRawParams) error {
	rawModels, err := b.CreateAlarmRawInfo(ctx, params)
	if !types.IsNil(err) {
		return err
	}

	// 查询策略
	strategy, err := b.strategyRepository.GetTeamStrategy(ctx, &bo.GetTeamStrategyParams{TeamID: params.TeamID, StrategyID: params.StrategyID})
	if !types.IsNil(err) {
		return err
	}

	// 查询策略等级
	level, err := b.strategyRepository.GetTeamStrategyLevel(ctx, &bo.GetTeamStrategyLevelParams{TeamID: params.TeamID, LevelID: params.LevelID})
	if !types.IsNil(err) {
		return err
	}

	// 查询告警列表数据源
	datasourceIds := types.SliceTo(params.Alerts, func(item *bo.AlertItemRawParams) uint32 {
		if types.IsNil(item.Labels) {
			return 0
		}
		return vobj.NewLabels(item.Labels).GetDatasourceID()
	})

	datasourceList, err := b.datasourceRepository.GetTeamDatasource(ctx, params.TeamID, datasourceIds)
	if !types.IsNil(err) {
		return err
	}

	datasourceMap := types.ToMap(datasourceList, func(datasource *bizmodel.Datasource) uint32 { return datasource.ID })

	rawInfoMap := types.ToMap(rawModels, func(rawModel *alarmmodel.AlarmRaw) string { return rawModel.Fingerprint })

	return b.SaveAlarmInfoDB(ctx,
		&bo.CreateAlarmInfoParams{
			RawInfoMap:    rawInfoMap,
			TeamID:        params.TeamID,
			Strategy:      strategy,
			Level:         level,
			Alerts:        params.Alerts,
			DatasourceMap: datasourceMap,
		})
}

// SaveAlarmInfoDB 保存告警信息db(告警历史、实时告警)
func (b *AlarmBiz) SaveAlarmInfoDB(ctx context.Context, params *bo.CreateAlarmInfoParams) error {
	// 保存告警历史
	if err := b.historyRepository.CreateAlarmHistory(ctx, params); !types.IsNil(err) {
		return err
	}

	// 保存实时告警
	if err := b.alarmRepository.CreateRealTimeAlarm(ctx, params); !types.IsNil(err) {
		return err
	}

	// 发送告警
	if err := b.sendAlert.Send(ctx, params.RawInfoMap); !types.IsNil(err) {
		return err
	}
	return nil
}
