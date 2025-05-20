package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/go-kratos/kratos/v2/log"
)

func NewTeamHookBiz(
	teamHookRepo repository.TeamHook,
	logger log.Logger,
) *TeamHook {
	return &TeamHook{
		helper:       log.NewHelper(log.With(logger, "module", "biz.team_hook")),
		teamHookRepo: teamHookRepo,
	}
}

type TeamHook struct {
	helper       *log.Helper
	teamHookRepo repository.TeamHook
}

// SaveHook 保存团队通知钩子
func (h *TeamHook) SaveHook(ctx context.Context, req *bo.SaveTeamNoticeHookRequest) error {
	if req.GetID() <= 0 {
		return h.teamHookRepo.Create(ctx, req)
	}
	hookDo, err := h.teamHookRepo.Get(ctx, req.GetID())
	if err != nil {
		return err
	}
	hook := req.WithUpdateHookRequest(hookDo)
	return h.teamHookRepo.Update(ctx, hook)
}

// UpdateHookStatus 更新钩子状态
func (h *TeamHook) UpdateHookStatus(ctx context.Context, req *bo.UpdateTeamNoticeHookStatusRequest) error {
	if req.HookID <= 0 {
		return merr.ErrorParams("invalid hook id")
	}
	return h.teamHookRepo.UpdateStatus(ctx, req)
}

// DeleteHook 删除钩子
func (h *TeamHook) DeleteHook(ctx context.Context, hookID uint32) error {
	if hookID <= 0 {
		return merr.ErrorParams("invalid hook id")
	}
	return h.teamHookRepo.Delete(ctx, hookID)
}

// GetHook 获取钩子详情
func (h *TeamHook) GetHook(ctx context.Context, hookID uint32) (do.NoticeHook, error) {
	if hookID <= 0 {
		return nil, merr.ErrorParams("invalid hook id")
	}
	return h.teamHookRepo.Get(ctx, hookID)
}

// ListHook 获取钩子列表
func (h *TeamHook) ListHook(ctx context.Context, req *bo.ListTeamNoticeHookRequest) (*bo.ListTeamNoticeHookReply, error) {
	if req == nil {
		return nil, merr.ErrorParams("invalid request")
	}
	return h.teamHookRepo.List(ctx, req)
}

// SelectHook 获取钩子列表
func (h *TeamHook) SelectHook(ctx context.Context, req *bo.TeamNoticeHookSelectRequest) (*bo.TeamNoticeHookSelectReply, error) {
	if req == nil {
		return nil, merr.ErrorParams("invalid request")
	}
	return h.teamHookRepo.Select(ctx, req)
}
