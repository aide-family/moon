package alarmhistory

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/model/historyscopes"
	"prometheus-manager/pkg/util/slices"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
)

var _ repository.HistoryRepo = (*alarmHistoryRepoImpl)(nil)

type alarmHistoryRepoImpl struct {
	repository.UnimplementedHistoryRepo
	data *data.Data

	log *log.Helper
	query.IAction[model.PromAlarmHistory]
}

func (l *alarmHistoryRepoImpl) GetHistoryById(ctx context.Context, id uint32) (*bo.AlarmHistoryBO, error) {
	detail, err := l.WithContext(ctx).FirstByID(id)
	if err != nil {
		return nil, err
	}

	return bo.AlarmHistoryModelToBO(detail), nil
}

func (l *alarmHistoryRepoImpl) ListHistory(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.AlarmHistoryBO, error) {
	list, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	boList := slices.To(list, func(v *model.PromAlarmHistory) *bo.AlarmHistoryBO {
		return bo.AlarmHistoryModelToBO(v)
	})
	return boList, nil
}

func (l *alarmHistoryRepoImpl) StorageHistory(ctx context.Context, historyBOs ...*bo.AlarmHistoryBO) ([]*bo.AlarmHistoryBO, error) {
	md5s := make([]string, 0, len(historyBOs))
	newModels := slices.To(historyBOs, func(v *bo.AlarmHistoryBO) *model.PromAlarmHistory {
		md5s = append(md5s, v.Md5)
		return v.ToModel()
	})
	if err := l.DB().WithContext(ctx).Scopes(historyscopes.ClausesOnConflict()).CreateInBatches(newModels, 50).Error; err != nil {
		return nil, err
	}

	var historyList []*model.PromAlarmHistory
	if err := l.DB().WithContext(ctx).Scopes(historyscopes.WhereInMd5(md5s...)).Find(&historyList).Error; err != nil {
		return nil, err
	}

	resList := slices.To(historyList, func(v *model.PromAlarmHistory) *bo.AlarmHistoryBO {
		return bo.AlarmHistoryModelToBO(v)
	})
	return resList, nil
}

func (l *alarmHistoryRepoImpl) UpdateHistoryById(ctx context.Context, id uint32, historyBO *bo.AlarmHistoryBO) (*bo.AlarmHistoryBO, error) {
	newModel := historyBO.ToModel()
	if err := l.WithContext(ctx).Scopes(historyscopes.ClausesOnConflict()).UpdateByID(id, newModel); err != nil {
		return nil, err
	}
	return bo.AlarmHistoryModelToBO(newModel), nil
}

// NewAlarmHistoryRepo .
func NewAlarmHistoryRepo(d *data.Data, logger log.Logger) repository.HistoryRepo {
	return &alarmHistoryRepoImpl{
		data: d,
		log:  log.NewHelper(log.With(logger, "module", "data.alarmHistoryRepoImpl")),
		IAction: query.NewAction[model.PromAlarmHistory](
			query.WithDB[model.PromAlarmHistory](d.DB()),
		),
	}
}
