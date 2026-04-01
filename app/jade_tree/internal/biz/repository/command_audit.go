package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/jade_tree/internal/biz/bo"
)

// CommandAudit stores and reviews SSH command proposals.
type CommandAudit interface {
	Create(ctx context.Context, in *bo.CommandAuditCreateRepoBo) (*bo.SSHCommandAuditItemBo, error)
	Get(ctx context.Context, uid snowflake.ID) (*bo.SSHCommandAuditItemBo, error)
	List(ctx context.Context, req *bo.ListSSHCommandAuditsBo) (*bo.PageResponseBo[*bo.SSHCommandAuditItemBo], error)
	Reject(ctx context.Context, in *bo.CommandAuditRejectBo) (*bo.SSHCommandAuditItemBo, error)
	Approve(ctx context.Context, uid, reviewer snowflake.ID) (*bo.SSHCommandAuditItemBo, *bo.SSHCommandItemBo, error)
}
