package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"

	pb "prometheus-manager/api/alarm/history"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/pkg/util/slices"
)

type (

	// HistoryBiz .
	HistoryBiz struct {
		log *log.Helper

		historyRepo      repository.HistoryRepo
		msgRepo          repository.MsgRepo
		strategyRepo     repository.StrategyRepo
		alarmRealtimeBiz *AlarmRealtimeBiz
		logX             repository.SysLogRepo
	}
)

// NewHistoryBiz .
func NewHistoryBiz(
	historyRepo repository.HistoryRepo,
	msgRepo repository.MsgRepo,
	strategyRepo repository.StrategyRepo,
	alarmRealtimeBiz *AlarmRealtimeBiz,
	logX repository.SysLogRepo,
	logger log.Logger,
) *HistoryBiz {
	return &HistoryBiz{
		log: log.NewHelper(log.With(logger, "module", "biz.alarmHistory")),

		historyRepo:      historyRepo,
		msgRepo:          msgRepo,
		strategyRepo:     strategyRepo,
		alarmRealtimeBiz: alarmRealtimeBiz,
		logX:             logX,
	}
}

// GetHistoryDetail 查询历史详情
func (a *HistoryBiz) GetHistoryDetail(ctx context.Context, id uint32) (*bo.AlarmHistoryBO, error) {
	historyDetail, err := a.historyRepo.GetHistoryById(ctx, id)
	if err != nil {
		return nil, err
	}
	return historyDetail, nil
}

// ListHistory 查询历史列表
func (a *HistoryBiz) ListHistory(ctx context.Context, req *pb.ListHistoryRequest) ([]*bo.AlarmHistoryBO, basescopes.Pagination, error) {
	pgReq := req.GetPage()
	pgInfo := basescopes.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	scopes := []basescopes.ScopeMethod{
		basescopes.LikeInstance(req.GetKeyword()),
		basescopes.TimeRange(req.GetStartAt(), req.GetEndAt()),
		basescopes.PreloadRealtimeAssociationStrategy(),
		basescopes.PreloadLevel(),
		basescopes.UpdateAtDesc(),
		basescopes.CreatedAtDesc(),
	}
	historyList, err := a.historyRepo.ListHistory(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, nil, err
	}
	return historyList, pgInfo, nil
}

// HandleHistory 维护告警数据
func (a *HistoryBiz) HandleHistory(ctx context.Context, hookBytes []byte, historyBO ...*bo.AlarmHistoryBO) ([]*bo.AlarmHistoryBO, error) {
	if len(historyBO) == 0 {
		return nil, nil
	}

	// 创建历史记录 or 更新历史记录
	historyBos, err := a.historyRepo.StorageHistory(ctx, historyBO...)
	if err != nil {
		return nil, err
	}

	realtimeAlarmBOs := slices.To(historyBos, func(alarmHistoryBO *bo.AlarmHistoryBO) *bo.AlarmRealtimeBO {
		return alarmHistoryBO.NewAlarmRealtimeBO()
	})

	// 处理实时告警
	realtimeAlarmBOs, err = a.alarmRealtimeBiz.HandleRealtime(ctx, realtimeAlarmBOs...)
	if err != nil {
		return nil, err
	}
	strategyIds := slices.To(historyBos, func(alarmHistoryBO *bo.AlarmHistoryBO) uint32 {
		return alarmHistoryBO.StrategyId
	})
	// 通过策略ID查询策略及下属通知对象信息
	wheres := []basescopes.ScopeMethod{
		basescopes.InIds(strategyIds...),
		basescopes.StrategyTablePreloadPromNotifies(
			basescopes.NotifyTablePreloadKeyChatGroups,
			basescopes.NotifyTablePreloadKeyBeNotifyMembers,
		),
	}
	strategyBOs, err := a.strategyRepo.List(ctx, wheres...)
	if err != nil {
		return nil, err
	}
	strategyBOsMap := make(map[uint32]*bo.StrategyBO)
	for _, strategyBO := range strategyBOs {
		strategyBOsMap[strategyBO.Id] = strategyBO
	}

	// TODO 构建告警消息
	alarmMsgList := make([]*bo.AlarmMsgBo, 0, len(historyBos))
	for _, historyBOItem := range historyBos {
		var promNotifies []*bo.NotifyBO
		if strategyBO, ok := strategyBOsMap[historyBOItem.StrategyId]; ok {
			promNotifies = strategyBO.GetPromNotifies()
		}
		alarmMsgList = append(alarmMsgList, &bo.AlarmMsgBo{
			AlarmStatus:  historyBOItem.Status,
			AlarmInfo:    historyBOItem.Info,
			StartsAt:     historyBOItem.StartsAt,
			EndsAt:       historyBOItem.EndsAt,
			StrategyBO:   strategyBOsMap[historyBOItem.StrategyId],
			PromNotifies: promNotifies,
		})
	}
	// 发送告警
	if err = a.msgRepo.SendAlarm(ctx, hookBytes, alarmMsgList...); err != nil {
		return nil, err
	}

	return historyBos, nil
}
