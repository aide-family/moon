package repository

import (
    "context"

    "github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
    "github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
)

type TeamSMSConfig interface {
    Get(ctx context.Context, id uint32) (do.TeamSMSConfig, error)
    List(ctx context.Context, req *bo.ListSMSConfigRequest) (*bo.ListSMSConfigListReply, error)
    Create(ctx context.Context, config bo.TeamSMSConfig) error
    Update(ctx context.Context, config bo.TeamSMSConfig) error
}