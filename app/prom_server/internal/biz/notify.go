package biz

import (
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"prometheus-manager/api/perrors"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
)

type NotifyBiz struct {
	log *log.Helper

	notifyRepo repository.NotifyRepo
}

func NewNotifyBiz(repo repository.NotifyRepo, logger log.Logger) *NotifyBiz {
	return &NotifyBiz{
		log:        log.NewHelper(log.With(logger, "module", "biz.NotifyBiz")),
		notifyRepo: repo,
	}
}

// CreateNotify 创建通知对象
func (b *NotifyBiz) CreateNotify(ctx context.Context, notifyBo *bo.NotifyBO) (*bo.NotifyBO, error) {
	notifyBo, err := b.notifyRepo.Create(ctx, notifyBo)
	if err != nil {
		return nil, err
	}

	return notifyBo, nil
}

// CheckNotifyName 检查通知名称是否存在
func (b *NotifyBiz) CheckNotifyName(ctx context.Context, name string, id ...uint32) error {
	total, err := b.notifyRepo.Count(ctx, basescopes.NameEQ(name), basescopes.NotInIds(id...))
	if err != nil {
		return err
	}
	if total > 0 {
		return perrors.ErrorAlreadyExists("通知对象名称已存在")
	}

	return nil
}

// UpdateNotifyById 更新通知对象
func (b *NotifyBiz) UpdateNotifyById(ctx context.Context, id uint32, notifyBo *bo.NotifyBO) error {
	return b.notifyRepo.Update(ctx, notifyBo, basescopes.InIds(id))
}

// DeleteNotifyById 删除通知对象
func (b *NotifyBiz) DeleteNotifyById(ctx context.Context, id uint32) error {
	return b.notifyRepo.Delete(ctx, basescopes.InIds(id))
}

// GetNotifyById 获取通知对象
func (b *NotifyBiz) GetNotifyById(ctx context.Context, id uint32) (*bo.NotifyBO, error) {
	wheres := []basescopes.ScopeMethod{
		basescopes.InIds(id),
		basescopes.NotifyPreloadChatGroups(),
		basescopes.NotifyPreloadBeNotifyMembers(),
	}
	notifyBo, err := b.notifyRepo.Get(ctx, wheres...)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, perrors.ErrorNotFound("通知对象不存在")
		}
		return nil, err
	}

	return notifyBo, nil
}

// ListNotify 获取通知对象列表
func (b *NotifyBiz) ListNotify(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.NotifyBO, error) {
	notifyBos, err := b.notifyRepo.List(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	return notifyBos, nil
}
