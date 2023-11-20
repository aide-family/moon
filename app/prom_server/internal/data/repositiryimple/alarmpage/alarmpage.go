package alarmpage

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/model"
)

var _ repository.PageRepo = (*alarmPageRepoImpl)(nil)

type alarmPageRepoImpl struct {
	log *log.Helper

	data *data.Data
	query.IAction[model.PromAlarmPage]
}

func (l *alarmPageRepoImpl) CreatePage(ctx context.Context, pageDo *dobo.AlarmPageDO) (*dobo.AlarmPageDO, error) {
	newModel := dobo.PageDOToModel(pageDo)
	if err := l.WithContext(ctx).Create(newModel); err != nil {
		return nil, err
	}
	return dobo.PageModelToDO(newModel), nil
}

func (l *alarmPageRepoImpl) UpdatePageById(ctx context.Context, id uint, pageDo *dobo.AlarmPageDO) (*dobo.AlarmPageDO, error) {
	newModel := dobo.PageDOToModel(pageDo)
	if err := l.WithContext(ctx).UpdateByID(id, newModel); err != nil {
		return nil, err
	}
	return dobo.PageModelToDO(newModel), nil
}

func (l *alarmPageRepoImpl) BatchUpdatePageStatusByIds(ctx context.Context, status int32, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	if err := l.WithContext(ctx).Update(&model.PromAlarmPage{Status: status}, query.WhereInColumn("id", ids)); err != nil {
		return err
	}
	return nil
}

func (l *alarmPageRepoImpl) DeletePageByIds(ctx context.Context, ids ...uint) error {
	if len(ids) == 0 {
		return nil
	}

	if err := l.WithContext(ctx).Delete(query.WhereInColumn("id", ids)); err != nil {
		return err
	}
	return nil
}

func (l *alarmPageRepoImpl) GetPageById(ctx context.Context, id uint) (*dobo.AlarmPageDO, error) {
	detail, err := l.WithContext(ctx).FirstByID(id)
	if err != nil {
		return nil, err
	}
	return dobo.PageModelToDO(detail), nil
}

func (l *alarmPageRepoImpl) ListPage(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.AlarmPageDO, error) {
	list, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	doList := make([]*dobo.AlarmPageDO, 0, len(list))
	for _, m := range list {
		doList = append(doList, dobo.PageModelToDO(m))
	}
	return doList, nil
}

// NewAlarmPageRepo .
func NewAlarmPageRepo(d *data.Data, logger log.Logger) repository.PageRepo {
	return &alarmPageRepoImpl{
		data: d,
		log:  log.NewHelper(log.With(logger, "module", "data.alarmPage")),
		IAction: query.NewAction[model.PromAlarmPage](
			query.WithDB[model.PromAlarmPage](d.DB()),
		),
	}
}
