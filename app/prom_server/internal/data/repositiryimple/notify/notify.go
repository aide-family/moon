package notify

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api/perrors"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.NotifyRepo = (*notifyRepoImpl)(nil)

var (
	// ErrNoCondition 不允许无条件操作DB
	ErrNoCondition = perrors.ErrorUnknown("no condition")
)

type notifyRepoImpl struct {
	repository.UnimplementedNotifyRepo
	log  *log.Helper
	data *data.Data

	query.IAction[model.PromAlarmNotify]
}

func (l *notifyRepoImpl) Get(ctx context.Context, scopes ...query.ScopeMethod) (*dobo.NotifyDO, error) {
	notifyDetail, err := l.WithContext(ctx).First(scopes...)
	if err != nil {
		return nil, err
	}

	return dobo.NotifyModelToDO(notifyDetail), nil
}

func (l *notifyRepoImpl) Find(ctx context.Context, scopes ...query.ScopeMethod) ([]*dobo.NotifyDO, error) {
	var notifyList []*model.PromAlarmNotify
	if err := l.DB().WithContext(ctx).Scopes(scopes...).Find(&notifyList).Error; err != nil {
		return nil, err
	}

	return slices.To(notifyList, func(i *model.PromAlarmNotify) *dobo.NotifyDO {
		return dobo.NotifyModelToDO(i)
	}), nil
}

func (l *notifyRepoImpl) Count(ctx context.Context, scopes ...query.ScopeMethod) (int64, error) {
	return l.WithContext(ctx).Count(scopes...)
}

func (l *notifyRepoImpl) List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.NotifyDO, error) {
	notifyList, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	return slices.To(notifyList, func(i *model.PromAlarmNotify) *dobo.NotifyDO {
		return dobo.NotifyModelToDO(i)
	}), nil
}

func (l *notifyRepoImpl) Create(ctx context.Context, notify *dobo.NotifyDO) (*dobo.NotifyDO, error) {
	newNotify := notify.ToModel()
	if err := l.WithContext(ctx).Create(newNotify); err != nil {
		return nil, err
	}

	return dobo.NotifyModelToDO(newNotify), nil
}

func (l *notifyRepoImpl) Update(ctx context.Context, notify *dobo.NotifyDO, scopes ...query.ScopeMethod) error {
	if len(scopes) == 0 {
		return ErrNoCondition
	}
	return l.WithContext(ctx).Update(notify.ToModel(), scopes...)
}

func (l *notifyRepoImpl) Delete(ctx context.Context, scopes ...query.ScopeMethod) error {
	if len(scopes) == 0 {
		return ErrNoCondition
	}

	return l.WithContext(ctx).Delete(scopes...)
}

func NewNotifyRepo(d *data.Data, logger log.Logger) repository.NotifyRepo {
	return &notifyRepoImpl{
		log:     log.NewHelper(log.With(logger, "module", "data.repository.notify")),
		data:    d,
		IAction: query.NewAction[model.PromAlarmNotify](query.WithDB[model.PromAlarmNotify](d.DB())),
	}
}
