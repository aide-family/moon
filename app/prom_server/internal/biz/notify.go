package biz

import (
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/msg"

	"prometheus-manager/api/perrors"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
)

type NotifyBiz struct {
	log *log.Helper

	notifyRepo repository.NotifyRepo
	msgRepo    repository.MsgRepo
	logX       repository.SysLogRepo
}

func NewNotifyBiz(
	repo repository.NotifyRepo,
	logX repository.SysLogRepo,
	msgRepo repository.MsgRepo,
	logger log.Logger,
) *NotifyBiz {
	return &NotifyBiz{
		log:        log.NewHelper(log.With(logger, "module", "biz.NotifyBiz")),
		notifyRepo: repo,
		logX:       logX,
		msgRepo:    msgRepo,
	}
}

// CreateNotify 创建通知对象
func (b *NotifyBiz) CreateNotify(ctx context.Context, notifyBo *bo.NotifyBO) (*bo.NotifyBO, error) {
	notifyBo, err := b.notifyRepo.Create(ctx, notifyBo)
	if err != nil {
		return nil, err
	}

	b.logX.CreateSysLog(ctx, vobj.ActionCreate, &bo.SysLogBo{
		ModuleName: vobj.ModuleAlarmNotifyGroup,
		ModuleId:   notifyBo.Id,
		Content:    notifyBo.String(),
		Title:      "创建通知对象",
	})

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
	// 查询
	oldData, err := b.GetNotifyById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return perrors.ErrorNotFound("通知对象不存在")
		}
		return err
	}
	if err = b.notifyRepo.Update(ctx, notifyBo, basescopes.InIds(id)); err != nil {
		return err
	}
	b.logX.CreateSysLog(ctx, vobj.ActionUpdate, &bo.SysLogBo{
		ModuleName: vobj.ModuleAlarmNotifyGroup,
		ModuleId:   id,
		Content:    bo.NewChangeLogBo(oldData, notifyBo).String(),
		Title:      "更新通知对象",
	})
	return nil
}

// DeleteNotifyById 删除通知对象
func (b *NotifyBiz) DeleteNotifyById(ctx context.Context, id uint32) error {
	// 查询
	oldData, err := b.GetNotifyById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return perrors.ErrorNotFound("通知对象不存在")
		}
		return err
	}
	if err = b.notifyRepo.Delete(ctx, basescopes.InIds(id)); err != nil {
		return err
	}
	b.logX.CreateSysLog(ctx, vobj.ActionDelete, &bo.SysLogBo{
		ModuleName: vobj.ModuleAlarmNotifyGroup,
		ModuleId:   id,
		Content:    oldData.String(),
		Title:      "删除通知对象",
	})
	return nil
}

// GetNotifyById 获取通知对象
func (b *NotifyBiz) GetNotifyById(ctx context.Context, id uint32) (*bo.NotifyBO, error) {
	wheres := []basescopes.ScopeMethod{
		basescopes.InIds(id),
		do.PromAlarmNotifyPreloadChatGroups(),
		do.PromAlarmNotifyPreloadBeNotifyMembers(),
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
func (b *NotifyBiz) ListNotify(ctx context.Context, req *bo.ListNotifyRequest) ([]*bo.NotifyBO, error) {
	wheres := []basescopes.ScopeMethod{
		basescopes.NameLike(req.Keyword),
		basescopes.StatusEQ(req.Status),
		basescopes.UpdateAtDesc(),
		basescopes.CreatedAtDesc(),
		do.PromAlarmNotifyPreloadChatGroups(),
		do.PromAlarmNotifyPreloadBeNotifyMembers(),
	}
	notifyBos, err := b.notifyRepo.List(ctx, req.Page, wheres...)
	if err != nil {
		return nil, err
	}

	return notifyBos, nil
}

// SendAlarmMessage 发送告警消息
func (b *NotifyBiz) SendAlarmMessage(ctx context.Context, chatInfo *bo.ChatGroupBO, message *msg.HookNotifyMsg) error {
	return msg.NewHookNotify(chatInfo.NotifyApp).Alarm(ctx, chatInfo.Hook, message)
}
