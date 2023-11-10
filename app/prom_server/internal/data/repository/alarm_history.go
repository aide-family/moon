package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz"

	"prometheus-manager/app/prom_server/internal/biz/alarmbiz"
	"prometheus-manager/app/prom_server/internal/data"
)

type alarmHistoryRepo struct {
	data *data.Data

	log *log.Helper
}

func (l *alarmHistoryRepo) GetHistoryById(ctx context.Context, id uint) (*biz.AlarmHistoryDO, error) {
	//TODO implement me
	panic("implement me")
}

func (l *alarmHistoryRepo) ListHistory(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*biz.AlarmHistoryDO, error) {
	//TODO implement me
	panic("implement me")
}

func (l *alarmHistoryRepo) CreateHistory(ctx context.Context, historyDo *biz.AlarmHistoryDO) (*biz.AlarmHistoryDO, error) {
	//TODO implement me
	panic("implement me")
}

func (l *alarmHistoryRepo) UpdateHistoryById(ctx context.Context, id uint, historyDo *biz.AlarmHistoryDO) (*biz.AlarmHistoryDO, error) {
	//TODO implement me
	panic("implement me")
}

// NewAlarmHistoryRepo .
func NewAlarmHistoryRepo(d *data.Data, logger log.Logger) alarmbiz.HistoryRepo {
	return &alarmHistoryRepo{
		data: d,
		log:  log.NewHelper(log.With(logger, "module", "data.alarmHistory")),
	}
}
