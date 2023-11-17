package biz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/alarm/history"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/pkg/model/history"
)

type (

	// HistoryBiz .
	HistoryBiz struct {
		log *log.Helper

		historyRepo repository.HistoryRepo
	}
)

// NewHistoryBiz .
func NewHistoryBiz(historyRepo repository.HistoryRepo, logger log.Logger) *HistoryBiz {
	return &HistoryBiz{
		log: log.NewHelper(log.With(logger, "module", "biz.alarmHistory")),

		historyRepo: historyRepo,
	}
}

// GetHistoryDetail 查询历史详情
func (a *HistoryBiz) GetHistoryDetail(ctx context.Context, id uint32) (*dobo.AlarmHistoryBO, error) {
	historyDetail, err := a.historyRepo.GetHistoryById(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	return dobo.NewAlarmHistoryDO(historyDetail).BO().First(), nil
}

// ListHistory 查询历史列表
func (a *HistoryBiz) ListHistory(ctx context.Context, req *pb.ListHistoryRequest) ([]*dobo.AlarmHistoryBO, query.Pagination, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	scopes := []query.ScopeMethod{
		history.LikeInstance(req.GetKeyword()),
		history.TimeRange(req.GetStartAt(), req.GetEndAt()),
	}
	historyList, err := a.historyRepo.ListHistory(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, nil, err
	}
	return dobo.NewAlarmHistoryDO(historyList...).BO().List(), pgInfo, nil
}

// CreateHistory 创建历史
func (a *HistoryBiz) CreateHistory(ctx context.Context, historyBO *dobo.AlarmHistoryBO) (*dobo.AlarmHistoryBO, error) {
	historyDO := dobo.NewAlarmHistoryBO(historyBO).DO().First()

	historyDO, err := a.historyRepo.CreateHistory(ctx, historyDO)
	if err != nil {
		return nil, err
	}

	return dobo.NewAlarmHistoryDO(historyDO).BO().First(), nil
}

// UpdateHistory 更新历史
func (a *HistoryBiz) UpdateHistory(ctx context.Context, historyBO *dobo.AlarmHistoryBO) (*dobo.AlarmHistoryBO, error) {
	historyDO := dobo.NewAlarmHistoryBO(historyBO).DO().First()

	historyDO, err := a.historyRepo.UpdateHistoryById(ctx, historyDO.Id, historyDO)
	if err != nil {
		return nil, err
	}

	return dobo.NewAlarmHistoryDO(historyDO).BO().First(), nil
}
