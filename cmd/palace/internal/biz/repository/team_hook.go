package repository

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
)

type TeamHook interface {
	Create(ctx context.Context, hook bo.NoticeHook) error
	Update(ctx context.Context, hook bo.NoticeHook) error
	UpdateStatus(ctx context.Context, req *bo.UpdateTeamNoticeHookStatusRequest) error
	Delete(ctx context.Context, hookID uint32) error
	Get(ctx context.Context, hookID uint32) (do.NoticeHook, error)
	List(ctx context.Context, req *bo.ListTeamNoticeHookRequest) (*bo.ListTeamNoticeHookReply, error)
	Find(ctx context.Context, ids []uint32) ([]do.NoticeHook, error)
}
