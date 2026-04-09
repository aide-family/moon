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

// AgentCommandDispatcher sends batch execute requests to remote jade_tree agents.
type AgentCommandDispatcher interface {
	BatchExecute(ctx context.Context, endpoint string, req *bo.BatchExecuteSSHCommandsBo) ([]*bo.BatchExecuteSSHCommandItemBo, error)
}
