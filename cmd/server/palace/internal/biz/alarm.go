package biz

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

// NewAlarmBiz 创建告警相关业务逻辑
func NewAlarmBiz(alarmRepository repository.Alarm, alarmRawRepository repository.AlarmRaw, historyRepository repository.HistoryRepository) *AlarmBiz {
	return &AlarmBiz{
		alarmRepository:    alarmRepository,
		alarmRawRepository: alarmRawRepository,
		historyRepository:  historyRepository,
	}
}

// AlarmBiz 告警相关业务逻辑
type AlarmBiz struct {
	alarmRepository    repository.Alarm
	alarmRawRepository repository.AlarmRaw
	historyRepository  repository.HistoryRepository
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
func (b *AlarmBiz) CreateAlarmRawInfo(ctx context.Context, param *bo.CreateAlarmHookRawParams) (*alarmmodel.AlarmRaw, error) {
	alarmRawJson, err := json.Marshal(param)
	if !types.IsNil(err) {
		return nil, err
	}

	return b.alarmRawRepository.CreateAlarmRaw(
		ctx,
		&bo.CreateAlarmRawParams{
			RawInfo:     string(alarmRawJson),
			Fingerprint: param.Fingerprint,
			TeamID:      param.TeamID,
		})
}

func (b *AlarmBiz) CreateAlarmInfo(ctx context.Context, params *bo.CreateAlarmHookRawParams) error {
	rawInfo, err := b.CreateAlarmRawInfo(ctx, params)
	if !types.IsNil(err) {
		return err
	}

	// 查询策略
	strategy, err := b.alarmRawRepository.GetTeamStrategy(ctx, &bo.GetTeamStrategyParams{TeamID: params.TeamID, StrategyID: params.StrategyID})
	if !types.IsNil(err) {
		return err
	}

	// 查询策略等级
	level, err := b.alarmRawRepository.GetStrategyLevel(ctx, &bo.GetTeamStrategyLevelParams{TeamID: params.TeamID, LevelID: params.LevelID})
	if !types.IsNil(err) {
		return err
	}

	// 查询告警列表数据源
	datasourceIds := types.SliceTo(params.Alerts, func(item *bo.CreateAlarmItemParams) uint32 {
		return item.DatasourceID
	})

	datasourceList, err := b.alarmRawRepository.ListDatasource(ctx, &bo.GetTeamDatasourceParams{DatasourceIds: datasourceIds, TeamID: params.TeamID})
	if !types.IsNil(err) {
		return err
	}

	datasourceMap := types.ToMap(datasourceList, func(datasource *bizmodel.Datasource) uint32 { return datasource.ID })

	return b.SaveAlarmInfoDB(ctx,
		&bo.CreateAlarmInfoParams{
			RawInfoID:     rawInfo.ID,
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
	return nil
}
