package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api/perrors"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
)

type (
	ChatGroupBiz struct {
		log *log.Helper

		chatGroupRepo repository.ChatGroupRepo
		logX          repository.SysLogRepo
	}
)

// NewChatGroupBiz .
func NewChatGroupBiz(chatGroupRepo repository.ChatGroupRepo, logX repository.SysLogRepo, logger log.Logger) *ChatGroupBiz {
	return &ChatGroupBiz{
		log:           log.NewHelper(logger),
		chatGroupRepo: chatGroupRepo,
		logX:          logX,
	}
}

// CreateChatGroup 创建通知群机器人hook
func (b *ChatGroupBiz) CreateChatGroup(ctx context.Context, chatGroup *bo.ChatGroupBO) (*bo.ChatGroupBO, error) {
	if chatGroup == nil {
		return nil, perrors.ErrorInvalidParams("参数错误")
	}
	newData, err := b.chatGroupRepo.Create(ctx, chatGroup)
	if err != nil {
		return nil, err
	}
	b.logX.CreateSysLog(ctx, vobj.ActionCreate, &bo.SysLogBo{
		ModuleName: vobj.ModuleAlarmNotifyHook,
		ModuleId:   newData.Id,
		Content:    newData.String(),
		Title:      "创建机器人hook",
	})
	return newData, nil
}

// GetChatGroupById  获取通知群机器人hook
func (b *ChatGroupBiz) GetChatGroupById(ctx context.Context, id uint32) (*bo.ChatGroupBO, error) {
	return b.chatGroupRepo.Get(ctx, basescopes.InIds(id))
}

// ListChatGroup 获取通知群机器人hook列表
func (b *ChatGroupBiz) ListChatGroup(ctx context.Context, req *bo.ListChatGroupReq) ([]*bo.ChatGroupBO, error) {
	wheres := []basescopes.ScopeMethod{
		basescopes.NameLike(req.Keyword),
		basescopes.StatusEQ(req.Status),
		basescopes.CreatedAtDesc(),
		basescopes.UpdateAtDesc(),
		do.PromAlarmChatGroupInApp(req.App),
		do.PromAlarmChatGroupPreloadCreateBy(),
	}
	return b.chatGroupRepo.List(ctx, req.Page, wheres...)
}

// UpdateChatGroupById 更新通知群机器人hook
func (b *ChatGroupBiz) UpdateChatGroupById(ctx context.Context, chatGroup *bo.ChatGroupBO, id uint32) error {
	// 查询详情
	chatGroupDetail, err := b.chatGroupRepo.Get(ctx, basescopes.InIds(id))
	if err != nil {
		return err
	}

	if err = b.chatGroupRepo.Update(ctx, chatGroup, basescopes.InIds(id)); err != nil {
		return err
	}

	b.logX.CreateSysLog(ctx, vobj.ActionUpdate, &bo.SysLogBo{
		ModuleName: vobj.ModuleAlarmNotifyHook,
		ModuleId:   id,
		Content:    bo.NewChangeLogBo(chatGroupDetail, chatGroup).String(),
		Title:      "更新机器人hook",
	})
	return nil
}

// DeleteChatGroupById 删除通知群机器人hook
func (b *ChatGroupBiz) DeleteChatGroupById(ctx context.Context, id uint32) error {
	if err := b.chatGroupRepo.Delete(ctx, basescopes.InIds(id)); err != nil {
		return err
	}
	b.logX.CreateSysLog(ctx, vobj.ActionDelete, &bo.SysLogBo{
		ModuleName: vobj.ModuleAlarmNotifyHook,
		ModuleId:   id,
		Title:      "删除机器人hook",
	})
	return nil
}
