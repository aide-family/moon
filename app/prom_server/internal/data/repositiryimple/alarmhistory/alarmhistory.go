package alarmhistory

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/model/history"

	"prometheus-manager/app/prom_server/internal/biz/dobo"
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

func (l *alarmHistoryRepoImpl) GetHistoryById(ctx context.Context, id uint) (*dobo.AlarmHistoryDO, error) {
	detail, err := l.WithContext(ctx).FirstByID(id)
	if err != nil {
		return nil, err
	}

	return historyModelToDO(detail), nil
}

func (l *alarmHistoryRepoImpl) ListHistory(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.AlarmHistoryDO, error) {
	list, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	boList := make([]*dobo.AlarmHistoryDO, 0, len(list))
	for _, v := range list {
		boList = append(boList, historyModelToDO(v))
	}
	return boList, nil
}

func (l *alarmHistoryRepoImpl) CreateHistory(ctx context.Context, historyDos ...*dobo.AlarmHistoryDO) ([]*dobo.AlarmHistoryDO, error) {
	newModels := make([]*model.PromAlarmHistory, 0, len(historyDos))
	for _, historyDo := range historyDos {
		newModel := historyDOToModel(historyDo)
		newModels = append(newModels, newModel)
	}
	if err := l.WithContext(ctx).Scopes(history.ClausesOnConflict()).BatchCreate(newModels, 50); err != nil {
		return nil, err
	}

	resList := make([]*dobo.AlarmHistoryDO, 0, len(newModels))
	for _, v := range newModels {
		resList = append(resList, historyModelToDO(v))
	}
	return resList, nil
}

func (l *alarmHistoryRepoImpl) UpdateHistoryById(ctx context.Context, id uint, historyDo *dobo.AlarmHistoryDO) (*dobo.AlarmHistoryDO, error) {
	newModel := historyDOToModel(historyDo)
	if err := l.WithContext(ctx).Scopes(history.ClausesOnConflict()).UpdateByID(id, newModel); err != nil {
		return nil, err
	}
	return historyModelToDO(newModel), nil
}

// historyModelToDO .
func historyModelToDO(detail *model.PromAlarmHistory) *dobo.AlarmHistoryDO {
	if detail == nil {
		return nil
	}
	return &dobo.AlarmHistoryDO{
		Id:         detail.ID,
		Md5:        detail.Md5,
		StrategyId: detail.StrategyID,
		LevelId:    detail.LevelID,
		Status:     detail.Status,
		StartAt:    detail.StartAt,
		EndAt:      detail.EndAt,
		Instance:   detail.Instance,
		Duration:   detail.Duration,
		Info:       detail.Info,
		CreatedAt:  detail.CreatedAt,
		UpdatedAt:  detail.UpdatedAt,
	}
}

// historyDOToModel .
func historyDOToModel(detail *dobo.AlarmHistoryDO) *model.PromAlarmHistory {
	if detail == nil {
		return nil
	}
	return &model.PromAlarmHistory{
		BaseModel: query.BaseModel{
			ID: detail.Id,
		},
		Instance:   detail.Instance,
		Status:     detail.Status,
		Info:       detail.Info,
		StartAt:    detail.StartAt,
		EndAt:      detail.EndAt,
		Duration:   detail.Duration,
		StrategyID: detail.StrategyId,
		LevelID:    detail.LevelId,
		Md5:        detail.Md5,
		Pages:      nil,
	}
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
