package notify

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/api/perrors"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
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
}

func (l *notifyRepoImpl) Get(ctx context.Context, scopes ...query.ScopeMethod) (*bo.NotifyBO, error) {
	var notifyDetail do.PromAlarmNotify
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).First(&notifyDetail).Error; err != nil {
		return nil, err
	}

	return bo.NotifyModelToBO(&notifyDetail), nil
}

func (l *notifyRepoImpl) Find(ctx context.Context, scopes ...query.ScopeMethod) ([]*bo.NotifyBO, error) {
	var notifyList []*do.PromAlarmNotify
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Find(&notifyList).Error; err != nil {
		return nil, err
	}

	return slices.To(notifyList, func(i *do.PromAlarmNotify) *bo.NotifyBO {
		return bo.NotifyModelToBO(i)
	}), nil
}

func (l *notifyRepoImpl) Count(ctx context.Context, scopes ...query.ScopeMethod) (int64, error) {
	var total int64
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (l *notifyRepoImpl) List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.NotifyBO, error) {
	var notifyList []*do.PromAlarmNotify
	if err := l.data.DB().WithContext(ctx).Scopes(append(scopes, basescopes.Page(pgInfo))...).Find(&notifyList).Error; err != nil {
		return nil, err
	}
	if pgInfo != nil {
		var total int64
		if err := l.data.DB().WithContext(ctx).Model(&do.PromAlarmNotify{}).Scopes(scopes...).Count(&total).Error; err != nil {
			return nil, err
		}
		pgInfo.SetTotal(total)
	}

	return slices.To(notifyList, func(i *do.PromAlarmNotify) *bo.NotifyBO {
		return bo.NotifyModelToBO(i)
	}), nil
}

func (l *notifyRepoImpl) Create(ctx context.Context, notify *bo.NotifyBO) (*bo.NotifyBO, error) {
	newNotify := notify.ToModel()
	if err := l.data.DB().WithContext(ctx).Create(newNotify).Error; err != nil {
		return nil, err
	}

	return bo.NotifyModelToBO(newNotify), nil
}

func (l *notifyRepoImpl) Update(ctx context.Context, notify *bo.NotifyBO, scopes ...query.ScopeMethod) error {
	if len(scopes) == 0 {
		return ErrNoCondition
	}
	return l.data.DB().WithContext(ctx).Scopes(scopes...).Updates(notify.ToModel()).Error
}

func (l *notifyRepoImpl) Delete(ctx context.Context, scopes ...query.ScopeMethod) error {
	if len(scopes) == 0 {
		return ErrNoCondition
	}

	return l.data.DB().WithContext(ctx).Scopes(scopes...).Delete(&do.PromAlarmNotify{}).Error
}

func NewNotifyRepo(d *data.Data, logger log.Logger) repository.NotifyRepo {
	return &notifyRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "data.repository.notify")),
		data: d,
	}
}
