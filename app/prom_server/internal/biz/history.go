package biz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/alarm/history"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/pkg/helper/model/history"
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
		history.LikeInstance(req.GetKeyword()),
		history.TimeRange(req.GetStartAt(), req.GetEndAt()),
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

	historyBos, err := a.historyRepo.CreateHistory(ctx, historyBO...)
	if err != nil {
		return nil, err
	}

	// 发送告警

	return historyBos, nil
}
