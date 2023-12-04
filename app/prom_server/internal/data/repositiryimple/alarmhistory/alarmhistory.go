package alarmhistory

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/model/historyscopes"

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

func (l *alarmHistoryRepoImpl) GetHistoryById(ctx context.Context, id uint) (*bo.AlarmHistoryBO, error) {
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
	boList := make([]*bo.AlarmHistoryBO, 0, len(list))
	for _, v := range list {
		boList = append(boList, bo.AlarmHistoryModelToBO(v))
	}
	return boList, nil
}

func (l *alarmHistoryRepoImpl) StorageHistory(ctx context.Context, historyBOs ...*bo.AlarmHistoryBO) ([]*bo.AlarmHistoryBO, error) {
	newModels := make([]*model.PromAlarmHistory, 0, len(historyBOs))
	for _, historyBO := range historyBOs {
		newModel := historyBO.ToModel()
		newModels = append(newModels, newModel)
	}
	if err := l.WithContext(ctx).Scopes(historyscopes.ClausesOnConflict()).BatchCreate(newModels, 50); err != nil {
		return nil, err
	}

	resList := make([]*bo.AlarmHistoryBO, 0, len(newModels))
	for _, v := range newModels {
		resList = append(resList, bo.AlarmHistoryModelToBO(v))
	}
	return resList, nil
}

func (l *alarmHistoryRepoImpl) UpdateHistoryById(ctx context.Context, id uint, historyBO *bo.AlarmHistoryBO) (*bo.AlarmHistoryBO, error) {
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
