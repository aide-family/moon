package alarmhistory

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/pkg/after"
	"prometheus-manager/pkg/helper/prom"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.HistoryRepo = (*alarmHistoryRepoImpl)(nil)

type alarmHistoryRepoImpl struct {
	repository.UnimplementedHistoryRepo
	data *data.Data

	log *log.Helper
}

func (l *alarmHistoryRepoImpl) GetHistoryById(ctx context.Context, id uint32) (*bo.AlarmHistoryBO, error) {
	var detail do.PromAlarmHistory
	if err := l.data.DB().WithContext(ctx).First(&detail, id).Error; err != nil {
		return nil, err
	}

	return bo.AlarmHistoryModelToBO(&detail), nil
}

func (l *alarmHistoryRepoImpl) ListHistory(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.AlarmHistoryBO, error) {
	var list []*do.PromAlarmHistory
	if err := l.data.DB().WithContext(ctx).Scopes(append(scopes, basescopes.Page(pgInfo))...).Find(&list).Error; err != nil {
		return nil, err
	}
	if pgInfo != nil {
		var total int64
		if err := l.data.DB().WithContext(ctx).Model(&do.PromAlarmHistory{}).Count(&total).Error; err != nil {
			return nil, err
		}
		pgInfo.SetTotal(total)
	}
	boList := slices.To(list, func(v *do.PromAlarmHistory) *bo.AlarmHistoryBO {
		return bo.AlarmHistoryModelToBO(v)
	})
	return boList, nil
}

func (l *alarmHistoryRepoImpl) StorageHistory(ctx context.Context, historyBOs ...*bo.AlarmHistoryBO) ([]*bo.AlarmHistoryBO, error) {
	md5s := make([]string, 0, len(historyBOs))
	newModels := slices.To(historyBOs, func(v *bo.AlarmHistoryBO) *do.PromAlarmHistory {
		md5s = append(md5s, v.Md5)
		return v.ToModel()
	})

	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.ClausesOnConflict()).CreateInBatches(newModels, 50).Error; err != nil {
		return nil, err
	}

	strategyIds := make(map[string]float64, len(historyBOs))
	for _, v := range newModels {
		idStr := strconv.FormatUint(uint64(v.StrategyID), 10)
		strategyIds[idStr] += 1
	}

	// 告警事件计数
	go func() {
		defer after.Recover(l.log)
		for strategyId, count := range strategyIds {
			prom.AlarmEventCounter.WithLabelValues("strategy_id", strategyId).Add(count)
		}
	}()

	var historyList []*do.PromAlarmHistory
	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.WhereInMd5(md5s...)).Find(&historyList).Error; err != nil {
		return nil, err
	}

	resList := slices.To(historyList, func(v *do.PromAlarmHistory) *bo.AlarmHistoryBO {
		return bo.AlarmHistoryModelToBO(v)
	})
	return resList, nil
}

func (l *alarmHistoryRepoImpl) UpdateHistoryById(ctx context.Context, id uint32, historyBO *bo.AlarmHistoryBO) (*bo.AlarmHistoryBO, error) {
	newModel := historyBO.ToModel()
	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.ClausesOnConflict(), basescopes.InIds(id)).Updates(newModel).Error; err != nil {
		return nil, err
	}
	return bo.AlarmHistoryModelToBO(newModel), nil
}

// NewAlarmHistoryRepo .
func NewAlarmHistoryRepo(d *data.Data, logger log.Logger) repository.HistoryRepo {
	return &alarmHistoryRepoImpl{
		data: d,
		log:  log.NewHelper(log.With(logger, "module", "data.alarmHistoryRepoImpl")),
	}
}
