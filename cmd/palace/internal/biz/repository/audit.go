package repository

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
)

type Audit interface {
	Get(ctx context.Context, id uint32) (do.TeamAudit, error)
	TeamAuditList(ctx context.Context, req *bo.TeamAuditListRequest) (*bo.TeamAuditListReply, error)
	UpdateTeamAuditStatus(ctx context.Context, req bo.UpdateTeamAuditStatus) error
}
