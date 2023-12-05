package alarmpage

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
)

var _ repository.PageRepo = (*alarmPageRepoImpl)(nil)

type alarmPageRepoImpl struct {
	repository.UnimplementedPageRepo
	log *log.Helper

	data *data.Data
	query.IAction[model.PromAlarmPage]
}

func (l *alarmPageRepoImpl) CreatePage(ctx context.Context, pageBO *bo.AlarmPageBO) (*bo.AlarmPageBO, error) {
	newModel := pageBO.ToModel()
	if err := l.WithContext(ctx).Create(newModel); err != nil {
		return nil, err
	}
	return bo.AlarmPageModelToBO(newModel), nil
}

func (l *alarmPageRepoImpl) UpdatePageById(ctx context.Context, id uint32, pageBO *bo.AlarmPageBO) (*bo.AlarmPageBO, error) {
	newModel := pageBO.ToModel()
	if err := l.WithContext(ctx).UpdateByID(id, newModel); err != nil {
		return nil, err
	}
	return bo.AlarmPageModelToBO(newModel), nil
}

func (l *alarmPageRepoImpl) BatchUpdatePageStatusByIds(ctx context.Context, status valueobj.Status, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	if err := l.WithContext(ctx).Update(&model.PromAlarmPage{Status: status}, query.WhereInColumn("id", ids)); err != nil {
		return err
	}
	return nil
}

func (l *alarmPageRepoImpl) DeletePageByIds(ctx context.Context, ids ...uint32) error {
	if len(ids) == 0 {
		return nil
	}

	if err := l.WithContext(ctx).Delete(query.WhereInColumn("id", ids)); err != nil {
		return err
	}
	return nil
}

func (l *alarmPageRepoImpl) GetPageById(ctx context.Context, id uint32) (*bo.AlarmPageBO, error) {
	detail, err := l.WithContext(ctx).FirstByID(id)
	if err != nil {
		return nil, err
	}
	return bo.AlarmPageModelToBO(detail), nil
}

func (l *alarmPageRepoImpl) ListPage(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.AlarmPageBO, error) {
	list, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	doList := make([]*bo.AlarmPageBO, 0, len(list))
	for _, m := range list {
		doList = append(doList, bo.AlarmPageModelToBO(m))
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
