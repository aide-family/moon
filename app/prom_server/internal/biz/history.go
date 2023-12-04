package biz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/alarm/history"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/pkg/helper/model/historyscopes"
	"prometheus-manager/pkg/util/slices"
)

type (

	// HistoryBiz .
	HistoryBiz struct {
		log *log.Helper

		historyRepo      repository.HistoryRepo
		msgRepo          repository.MsgRepo
		alarmRealtimeBiz *AlarmRealtimeBiz
	}
)

// NewHistoryBiz .
func NewHistoryBiz(
	historyRepo repository.HistoryRepo,
	msgRepo repository.MsgRepo,
	alarmRealtimeBiz *AlarmRealtimeBiz,
	logger log.Logger,
) *HistoryBiz {
	return &HistoryBiz{
		log: log.NewHelper(log.With(logger, "module", "biz.alarmHistory")),

		historyRepo:      historyRepo,
		msgRepo:          msgRepo,
		alarmRealtimeBiz: alarmRealtimeBiz,
	}
}

// GetHistoryDetail 查询历史详情
func (a *HistoryBiz) GetHistoryDetail(ctx context.Context, id uint32) (*bo.AlarmHistoryBO, error) {
	historyDetail, err := a.historyRepo.GetHistoryById(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	return historyDetail, nil
}

// ListHistory 查询历史列表
func (a *HistoryBiz) ListHistory(ctx context.Context, req *pb.ListHistoryRequest) ([]*bo.AlarmHistoryBO, query.Pagination, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	scopes := []query.ScopeMethod{
		historyscopes.LikeInstance(req.GetKeyword()),
		historyscopes.TimeRange(req.GetStartAt(), req.GetEndAt()),
	}
	historyList, err := a.historyRepo.ListHistory(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, nil, err
	}
	return historyList, pgInfo, nil
}

// HandleHistory 维护告警数据
func (a *HistoryBiz) HandleHistory(ctx context.Context, historyBO ...*bo.AlarmHistoryBO) ([]*bo.AlarmHistoryBO, error) {
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

	// 发送告警
	if err = a.msgRepo.SendAlarm(ctx, realtimeAlarmBOs...); err != nil {
		return nil, err
	}

	return historyBos, nil
}
