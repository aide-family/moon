package alarmpage

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.PageRepo = (*alarmPageRepoImpl)(nil)

type alarmPageRepoImpl struct {
	repository.UnimplementedPageRepo
	log *log.Helper

	data *data.Data
}

func (l *alarmPageRepoImpl) GetPromStrategyAlarmPage(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*do.PromStrategyAlarmPage, error) {
	var m []*do.PromStrategyAlarmPage
	if err := l.data.DB().
		WithContext(ctx).
		Scopes(scopes...).
		Find(&m).
		Error; err != nil {
		return nil, err
	}
	return m, nil
}

func (l *alarmPageRepoImpl) GetStrategyIds(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]uint32, error) {
	var strategyIds []uint32
	if err := l.data.DB().
		Model(&do.PromStrategyAlarmPage{}).
		WithContext(ctx).
		Scopes(scopes...).
		Pluck(basescopes.TableNamePromStrategyAlarmPageFieldPromStrategyID.String(), &strategyIds).
		Error; err != nil {
		return nil, err
	}
	return strategyIds, nil
}

func (l *alarmPageRepoImpl) Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.AlarmPageBO, error) {
	var model do.PromAlarmPage
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).First(&model).Error; err != nil {
		return nil, err
	}

	return bo.AlarmPageModelToBO(&model), nil
}

func (l *alarmPageRepoImpl) CreatePage(ctx context.Context, pageBO *bo.AlarmPageBO) (*bo.AlarmPageBO, error) {
	newModel := pageBO.ToModel()
	if err := l.data.DB().WithContext(ctx).Create(newModel).Error; err != nil {
		return nil, err
	}
	return bo.AlarmPageModelToBO(newModel), nil
}

func (l *alarmPageRepoImpl) UpdatePageById(ctx context.Context, id uint32, pageBO *bo.AlarmPageBO) (*bo.AlarmPageBO, error) {
	newModel := pageBO.ToModel()
	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.InIds(id)).Updates(newModel).Error; err != nil {
		return nil, err
	}
	return bo.AlarmPageModelToBO(newModel), nil
}

func (l *alarmPageRepoImpl) BatchUpdatePageStatusByIds(ctx context.Context, status vo.Status, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.InIds(ids...)).Updates(&do.PromAlarmPage{Status: status}).Error; err != nil {
		return err
	}
	return nil
}

func (l *alarmPageRepoImpl) DeletePageByIds(ctx context.Context, ids ...uint32) error {
	if len(ids) == 0 {
		return nil
	}

	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.WhereInColumn("id", ids)).Delete(&do.PromAlarmPage{}).Error; err != nil {
		return err
	}
	return nil
}

func (l *alarmPageRepoImpl) GetPageById(ctx context.Context, id uint32) (*bo.AlarmPageBO, error) {
	var detail do.PromAlarmPage
	if err := l.data.DB().WithContext(ctx).First(&detail, id).Error; err != nil {
		return nil, err
	}
	return bo.AlarmPageModelToBO(&detail), nil
}

func (l *alarmPageRepoImpl) ListPage(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.AlarmPageBO, error) {
	var list []*do.PromAlarmPage
	if err := l.data.DB().WithContext(ctx).Scopes(append(scopes, basescopes.Page(pgInfo))...).Find(&list).Error; err != nil {
		return nil, err
	}
	if pgInfo != nil {
		var total int64
		if err := l.data.DB().WithContext(ctx).Model(&do.PromAlarmPage{}).Count(&total).Error; err != nil {
			return nil, err
		}
		pgInfo.SetTotal(total)
	}
	doList := slices.To(list, func(item *do.PromAlarmPage) *bo.AlarmPageBO {
		return bo.AlarmPageModelToBO(item)
	})
	return doList, nil
}

// NewAlarmPageRepo .
func NewAlarmPageRepo(d *data.Data, logger log.Logger) repository.PageRepo {
	return &alarmPageRepoImpl{
		data: d,
		log:  log.NewHelper(log.With(logger, "module", "data.alarmPage")),
	}
}
