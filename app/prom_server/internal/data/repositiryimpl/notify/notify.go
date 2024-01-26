package notify

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

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

func (l *notifyRepoImpl) Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.NotifyBO, error) {
	var notifyDetail do.PromAlarmNotify
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).First(&notifyDetail).Error; err != nil {
		return nil, err
	}

	return bo.NotifyModelToBO(&notifyDetail), nil
}

func (l *notifyRepoImpl) Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.NotifyBO, error) {
	var notifyList []*do.PromAlarmNotify
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Find(&notifyList).Error; err != nil {
		return nil, err
	}

	return slices.To(notifyList, func(i *do.PromAlarmNotify) *bo.NotifyBO {
		return bo.NotifyModelToBO(i)
	}), nil
}

func (l *notifyRepoImpl) Count(ctx context.Context, scopes ...basescopes.ScopeMethod) (int64, error) {
	var total int64
	if err := l.data.DB().WithContext(ctx).Model(&do.PromAlarmNotify{}).Scopes(scopes...).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (l *notifyRepoImpl) List(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.NotifyBO, error) {
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
	chatGroupModels := slices.To(notify.GetChatGroups(), func(i *bo.ChatGroupBO) *do.PromAlarmChatGroup { return i.ToModel() })
	notifyMembers := slices.To(notify.GetBeNotifyMembers(), func(i *bo.NotifyMemberBO) *do.PromAlarmNotifyMember { return i.ToModel() })
	err := l.data.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(newNotify).Create(newNotify).Error; err != nil {
			return err
		}
		if err := tx.Model(newNotify).Association(basescopes.NotifyTablePreloadKeyChatGroups).Replace(chatGroupModels); err != nil {
			return err
		}
		return tx.Model(newNotify).Association(basescopes.NotifyTablePreloadKeyBeNotifyMembers).Replace(notifyMembers)
	})
	if err != nil {
		return nil, err
	}

	return bo.NotifyModelToBO(newNotify), nil
}

func (l *notifyRepoImpl) Update(ctx context.Context, notify *bo.NotifyBO, scopes ...basescopes.ScopeMethod) error {
	if len(scopes) == 0 {
		return ErrNoCondition
	}
	newModel := notify.ToModel()
	chatGroupModels := slices.To(notify.GetChatGroups(), func(i *bo.ChatGroupBO) *do.PromAlarmChatGroup { return i.ToModel() })
	notifyMembers := slices.To(notify.GetBeNotifyMembers(), func(i *bo.NotifyMemberBO) *do.PromAlarmNotifyMember { return i.ToModel() })
	return l.data.DB().Model(newModel).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(newModel).Scopes(scopes...).Updates(newModel).Error; err != nil {
			l.log.Warnf("update notify error: %v", err)
			return err
		}
		if err := tx.Model(newModel).Association(basescopes.NotifyTablePreloadKeyChatGroups).Replace(chatGroupModels); err != nil {
			l.log.Warnf("update notify chat group error: %v", err)
			return err
		}
		if err := tx.Model(newModel).Association(basescopes.NotifyTablePreloadKeyBeNotifyMembers).Replace(notifyMembers); err != nil {
			l.log.Warnf("update notify member error: %v", err)
			return err
		}
		return nil
	})
}

func (l *notifyRepoImpl) Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error {
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
