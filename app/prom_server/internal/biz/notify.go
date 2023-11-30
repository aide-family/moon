package biz

import (
	"context"
	"errors"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"prometheus-manager/api/perrors"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/pkg/helper/model/notifyscopes"
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
func (b *NotifyBiz) CreateNotify(ctx context.Context, notifyBo *dobo.NotifyBO) (*dobo.NotifyBO, error) {
	notifyDo := dobo.NewNotifyBO(notifyBo).DO().First()
	notifyDo, err := b.notifyRepo.Create(ctx, notifyDo)
	if err != nil {
		return nil, err
	}

	return dobo.NewNotifyDO(notifyDo).BO().First(), nil
}

// CheckNotifyName 检查通知名称是否存在
func (b *NotifyBiz) CheckNotifyName(ctx context.Context, name string, id ...uint) error {
	total, err := b.notifyRepo.Count(ctx, notifyscopes.NotifyEqName(name), notifyscopes.NotifyNotInIds(id...))
	if err != nil {
		return err
	}
	if total > 0 {
		return perrors.ErrorAlreadyExists("通知对象名称已存在")
	}

	return nil
}

// UpdateNotifyById 更新通知对象
func (b *NotifyBiz) UpdateNotifyById(ctx context.Context, id uint, notifyBo *dobo.NotifyBO) error {
	notifyDo := dobo.NewNotifyBO(notifyBo).DO().First()
	return b.notifyRepo.Update(ctx, notifyDo, notifyscopes.NotifyInIds(id))
}

// DeleteNotifyById 删除通知对象
func (b *NotifyBiz) DeleteNotifyById(ctx context.Context, id uint) error {
	return b.notifyRepo.Delete(ctx, notifyscopes.NotifyInIds(id))
}

// GetNotifyById 获取通知对象
func (b *NotifyBiz) GetNotifyById(ctx context.Context, id uint) (*dobo.NotifyBO, error) {
	notifyDo, err := b.notifyRepo.Get(ctx, notifyscopes.NotifyInIds(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, perrors.ErrorNotFound("通知对象不存在")
		}
		return nil, err
	}

	return dobo.NewNotifyDO(notifyDo).BO().First(), nil
}

// ListNotify 获取通知对象列表
func (b *NotifyBiz) ListNotify(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.NotifyBO, error) {
	notifyDos, err := b.notifyRepo.List(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	return dobo.NewNotifyDO(notifyDos...).BO().List(), nil
}
