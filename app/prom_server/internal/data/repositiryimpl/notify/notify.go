package notify

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"prometheus-manager/pkg/helper/middler"

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
	whereList := append(scopes, basescopes.WithCreateBy(ctx))
	if err := l.data.DB().WithContext(ctx).Scopes(whereList...).First(&notifyDetail).Error; err != nil {
		return nil, err
	}

	return bo.NotifyModelToBO(&notifyDetail), nil
}

func (l *notifyRepoImpl) Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.NotifyBO, error) {
	var notifyList []*do.PromAlarmNotify
	whereList := append(scopes, basescopes.WithCreateBy(ctx))
	if err := l.data.DB().WithContext(ctx).Scopes(whereList...).Find(&notifyList).Error; err != nil {
		return nil, err
	}

	return slices.To(notifyList, func(i *do.PromAlarmNotify) *bo.NotifyBO {
		return bo.NotifyModelToBO(i)
	}), nil
}

func (l *notifyRepoImpl) Count(ctx context.Context, scopes ...basescopes.ScopeMethod) (int64, error) {
	var total int64
	whereList := append(scopes, basescopes.WithCreateBy(ctx))
	if err := l.data.DB().WithContext(ctx).Model(&do.PromAlarmNotify{}).Scopes(whereList...).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (l *notifyRepoImpl) List(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.NotifyBO, error) {
	var notifyList []*do.PromAlarmNotify
	whereList := append(scopes, basescopes.WithCreateBy(ctx))
	if err := l.data.DB().WithContext(ctx).Scopes(append(whereList, bo.Page(pgInfo))...).Find(&notifyList).Error; err != nil {
		return nil, err
	}
	if pgInfo != nil {
		var total int64
		if err := l.data.DB().WithContext(ctx).Model(&do.PromAlarmNotify{}).Scopes(whereList...).Count(&total).Error; err != nil {
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
	newNotify.CreateBy = middler.GetUserId(ctx)
	err := l.data.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(newNotify).Create(newNotify).Error; err != nil {
			return err
		}
		if err := tx.Model(newNotify).Association(do.PromAlarmNotifyPreloadFieldChatGroups).Replace(chatGroupModels); err != nil {
			return err
		}
		return nil
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
	return l.data.DB().Model(newModel).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(newModel).Scopes(append(scopes, basescopes.WithCreateBy(ctx))...).Updates(newModel).Error; err != nil {
			l.log.Warnf("update notify error: %v", err)
			return err
		}
		if err := tx.Model(newModel).Association(do.PromAlarmNotifyPreloadFieldChatGroups).Replace(chatGroupModels); err != nil {
			l.log.Warnf("update notify chat group error: %v", err)
			return err
		}
		// 删除旧的人员
		if err := tx.Model(&do.PromAlarmNotifyMember{}).Scopes(do.PromAlarmNotifyMemberWherePromAlarmNotifyID(notify.Id)).Delete(&do.PromAlarmNotifyMember{}).Error; err != nil {
			return err
		}
		notifyMembers := slices.To(notify.GetBeNotifyMembers(), func(i *bo.NotifyMemberBO) *do.PromAlarmNotifyMember {
			nm := i.ToModel()
			nm.PromAlarmNotifyID = newModel.ID
			return nm
		})

		return tx.Model(&do.PromAlarmNotifyMember{}).CreateInBatches(notifyMembers, 100).Error
	})
}

func (l *notifyRepoImpl) Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error {
	if len(scopes) == 0 {
		return ErrNoCondition
	}
	whereList := append(scopes, basescopes.WithCreateBy(ctx))
	return l.data.DB().WithContext(ctx).Scopes(whereList...).Delete(&do.PromAlarmNotify{}).Error
}

func NewNotifyRepo(d *data.Data, logger log.Logger) repository.NotifyRepo {
	return &notifyRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "data.repository.notify")),
		data: d,
	}
}
