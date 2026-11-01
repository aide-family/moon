package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// AgentRepo 读写机器与令牌哈希。
type AgentRepo interface {
	CreateAgent(ctx context.Context, in bo.Agent, tokenHash string) (*bo.Agent, error)
	UpdateAgent(ctx context.Context, in bo.Agent) (*bo.Agent, error)
	SetTokenHash(ctx context.Context, namespaceUID, uid, tokenHash, status string) (*bo.Agent, error)
	DeleteAgent(ctx context.Context, namespaceUID, uid string) error
	GetAgent(ctx context.Context, namespaceUID, uid string) (*bo.Agent, error)
	GetAgentByUID(ctx context.Context, uid string) (*bo.Agent, error)
	GetAgentByTokenHash(ctx context.Context, tokenHash string) (*bo.Agent, error)
	ListAgents(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.Agent, int, error)
	// MatchAgents 返回当前租户下标签全部命中的 ACTIVE 机器（AND 精确匹配）。
	MatchAgents(ctx context.Context, namespaceUID string, labels map[string]string) ([]bo.Agent, error)
	SetOnline(ctx context.Context, uid string, online bool, hostname, machineID, localIP, version, collectJSON string) error
	GetSSHSecrets(ctx context.Context, uid string) (password, privateKey string, err error)
}

// SSHCommandRepo 读写命令与审计。
type SSHCommandRepo interface {
	CreateCommand(ctx context.Context, in bo.SSHCommand, operatorUID string) (*bo.SSHCommand, error)
	UpdateCommandPending(ctx context.Context, in bo.SSHCommand, operatorUID string) (*bo.SSHCommand, error)
	ApproveCommand(ctx context.Context, namespaceUID, uid, operatorUID string) (*bo.SSHCommand, error)
	RejectCommand(ctx context.Context, namespaceUID, uid, operatorUID string) (*bo.SSHCommand, error)
	DeleteCommand(ctx context.Context, namespaceUID, uid string) error
	GetCommand(ctx context.Context, namespaceUID, uid string) (*bo.SSHCommand, error)
	ListCommands(ctx context.Context, namespaceUID string, page, pageSize int, name, reviewStatus string) ([]bo.SSHCommand, int, error)
	AddExecAudit(ctx context.Context, in bo.SSHCommandAudit) (*bo.SSHCommandAudit, error)
	ListAudits(ctx context.Context, namespaceUID string, query bo.SSHCommandAuditQuery, page, pageSize int) ([]bo.SSHCommandAudit, int, error)
}
