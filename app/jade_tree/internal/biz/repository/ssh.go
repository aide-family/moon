// Package repository declares data access contracts for biz logic.
package repository

import (
	"context"

	"github.com/aide-family/jade_tree/internal/biz/bo"
)

// SSHOperator executes commands on remote hosts over SSH.
type SSHOperator interface {
	Exec(ctx context.Context, req *bo.SSHExecRequest) (*bo.SSHExecReply, error)
}
