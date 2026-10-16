package repository

import (
	"context"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// EdgeAuditRepo 读写 SFTP 与 Kubernetes 代理审计。
type EdgeAuditRepo interface {
	CreateSFTPAudit(ctx context.Context, in bo.SFTPAudit) (*bo.SFTPAudit, error)
	UpdateSFTPAudit(ctx context.Context, uid, result, detail string) (*bo.SFTPAudit, error)
	ListSFTPAudits(ctx context.Context, namespaceUID, agentUID string, page, pageSize int) ([]bo.SFTPAudit, int, error)
	CreateK8sAudit(ctx context.Context, in bo.K8sProxyAudit) (*bo.K8sProxyAudit, error)
	UpdateK8sAudit(ctx context.Context, uid, result, detail string) (*bo.K8sProxyAudit, error)
	ListK8sAudits(ctx context.Context, namespaceUID, agentUID string, page, pageSize int) ([]bo.K8sProxyAudit, int, error)
}
