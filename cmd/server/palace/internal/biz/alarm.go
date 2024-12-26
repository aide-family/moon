package biz

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	hookapi "github.com/aide-family/moon/api/rabbit/hook"
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
	alarmGroup repository.AlarmGroup,
) *AlarmBiz {
	return &AlarmBiz{
		alarmGroup:           alarmGroup,
		alarmRepository:      alarmRepository,
		alarmRawRepository:   alarmRawRepository,
		historyRepository:    historyRepository,
		strategyRepository:   strategyRepository,
		datasourceRepository: datasourceRepository,
		sendAlertRepository:  sendAlert,
	}
}

// AlarmBiz 告警相关业务逻辑
type AlarmBiz struct {
	alarmGroup           repository.AlarmGroup
	alarmRepository      repository.Alarm
	alarmRawRepository   repository.AlarmRaw
	historyRepository    repository.HistoryRepository
	strategyRepository   repository.Strategy
	datasourceRepository repository.Datasource
	sendAlertRepository  microrepository.SendAlert
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

// CreateAlarmInfo 创建告警信息
func (b *AlarmBiz) CreateAlarmInfo(ctx context.Context, params *bo.CreateAlarmHookRawParams) error {
	rawModels, err := b.CreateAlarmRawInfo(ctx, params)
	if !types.IsNil(err) {
		return err
	}

	// 查询策略 TODO 增加缓存 2h
	strategy, err := b.strategyRepository.GetTeamStrategy(ctx, &bo.GetTeamStrategyParams{TeamID: params.TeamID, StrategyID: params.StrategyID})
	if err != nil {
		return err
	}

	level := strategy.Level.GetLevelByID(params.LevelID)

	// 查询告警列表数据源 TODO 增加缓存 2h
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
	receiverGroupIDs := types.SliceToWithFilter(strings.Split(params.Receiver, ","), func(item string) (uint32, bool) {
		ids := strings.Split(item, "_")
		if len(ids) == 0 {
			return 0, false
		}
		id, err := strconv.ParseUint(ids[len(ids)-1], 10, 32)
		return uint32(id), err == nil
	})
	return b.SaveAlarmInfoDB(ctx,
		&bo.CreateAlarmInfoParams{
			ReceiverGroupIDs: receiverGroupIDs,
			TeamID:           params.TeamID,
			Alerts:           params.Alerts,
			Strategy:         strategy,
			Level:            level,
			DatasourceMap:    datasourceMap,
			RawInfoMap:       rawInfoMap,
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
	if err := b.send(ctx, params); !types.IsNil(err) {
		return err
	}
	return nil
}

// send 发送告警
func (b *AlarmBiz) send(ctx context.Context, params *bo.CreateAlarmInfoParams) error {
	// 过滤不在允许条件内的告警
	alarmGroups, err := b.alarmGroup.GetAlarmGroupsByIDs(ctx, params.ReceiverGroupIDs)
	if !types.IsNil(err) {
		return err
	}

	for _, v := range params.Alerts {
		// 以告警发生时刻的时间作为时间引擎的判断时间
		ts := types.NewTimeByString(v.StartsAt)
		receivers := make([]string, 0, len(alarmGroups))
		for _, v := range alarmGroups {
			if !v.IsAllowed(ts.Time) {
				continue
			}
			receivers = append(receivers, fmt.Sprintf("team_%d_%d", params.TeamID, v.ID))
		}

		if len(receivers) == 0 {
			continue
		}
		for _, route := range receivers {
			key := v.NoticeKey(route)
			task := &hookapi.SendMsgRequest{
				Json:      v.GetAlertItemString(),
				Route:     route,
				RequestID: key,
			}
			b.SendAlertMsg(ctx, &bo.SendMsg{SendMsgRequest: task})
		}
	}
	return nil
}

// SendAlertMsg 发送告警消息
func (b *AlarmBiz) SendAlertMsg(ctx context.Context, params *bo.SendMsg) {
	b.sendAlertRepository.Send(ctx, params)
}
