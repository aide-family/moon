package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/alarmbiz"
	"prometheus-manager/app/prom_server/internal/data"
)

type alarmPageRepo struct {
	log *log.Helper

	data *data.Data
}

func (l *alarmPageRepo) CreatePage(ctx context.Context, pageDo *biz.AlarmPageDO) (*biz.AlarmPageDO, error) {
	//TODO implement me
	panic("implement me")
}

func (l *alarmPageRepo) UpdatePageById(ctx context.Context, id uint, pageDo *biz.AlarmPageDO) (*biz.AlarmPageDO, error) {
	//TODO implement me
	panic("implement me")
}

func (l *alarmPageRepo) BatchUpdatePageStatusByIds(ctx context.Context, status int32, ids []uint) error {
	//TODO implement me
	panic("implement me")
}

func (l *alarmPageRepo) DeletePageByIds(ctx context.Context, id ...uint) error {
	//TODO implement me
	panic("implement me")
}

func (l *alarmPageRepo) GetPageById(ctx context.Context, id uint) (*biz.AlarmPageDO, error) {
	//TODO implement me
	panic("implement me")
}

func (l *alarmPageRepo) ListPage(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*biz.AlarmPageDO, error) {
	//TODO implement me
	panic("implement me")
}

// NewAlarmPageRepo .
func NewAlarmPageRepo(d *data.Data, logger log.Logger) alarmbiz.PageRepo {
	return &alarmPageRepo{
		data: d,
		log:  log.NewHelper(log.With(logger, "module", "data.alarmPage")),
	}
}
