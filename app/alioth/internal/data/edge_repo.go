package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/k8sproxyaudit"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/sftpaudit"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/pkg/merr"
)

type edgeAuditRepo struct{ *repoBase }

func NewEdgeAuditRepo(d *Data, logger log.Logger) repository.EdgeAuditRepo {
	return &edgeAuditRepo{newBase(d, logger)}
}

func (r *edgeAuditRepo) CreateSFTPAudit(ctx context.Context, in bo.SFTPAudit) (*bo.SFTPAudit, error) {
	row, err := r.client.SFTPAudit.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetAgentUID(in.AgentUID).
		SetOperatorUID(in.OperatorUID).
		SetPath(in.Path).
		SetAction(in.Action).
		SetResult(in.Result).
		SetDetail(in.Detail).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toSFTPAudit(row)
	return &item, nil
}

func (r *edgeAuditRepo) UpdateSFTPAudit(ctx context.Context, uid, result, detail string) (*bo.SFTPAudit, error) {
	n, err := r.client.SFTPAudit.Update().Where(sftpaudit.UIDEQ(uid)).SetResult(result).SetDetail(detail).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	row, err := r.client.SFTPAudit.Query().Where(sftpaudit.UIDEQ(uid)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toSFTPAudit(row)
	return &item, nil
}

func (r *edgeAuditRepo) ListSFTPAudits(ctx context.Context, namespaceUID, agentUID string, page, pageSize int) ([]bo.SFTPAudit, int, error) {
	query := r.client.SFTPAudit.Query().Where(sftpaudit.NamespaceUIDEQ(namespaceUID))
	if agentUID != "" {
		query = query.Where(sftpaudit.AgentUIDEQ(agentUID))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(sftpaudit.FieldCreatedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.SFTPAudit, 0, len(rows))
	for _, row := range rows {
		items = append(items, toSFTPAudit(row))
	}
	return items, total, nil
}

func (r *edgeAuditRepo) CreateK8sAudit(ctx context.Context, in bo.K8sProxyAudit) (*bo.K8sProxyAudit, error) {
	row, err := r.client.K8sProxyAudit.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetAgentUID(in.AgentUID).
		SetOperatorUID(in.OperatorUID).
		SetMethod(in.Method).
		SetResource(in.Resource).
		SetKubeNamespace(in.KubeNamespace).
		SetResult(in.Result).
		SetDetail(in.Detail).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toK8sAudit(row)
	return &item, nil
}

func (r *edgeAuditRepo) UpdateK8sAudit(ctx context.Context, uid, result, detail string) (*bo.K8sProxyAudit, error) {
	n, err := r.client.K8sProxyAudit.Update().Where(k8sproxyaudit.UIDEQ(uid)).SetResult(result).SetDetail(detail).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	row, err := r.client.K8sProxyAudit.Query().Where(k8sproxyaudit.UIDEQ(uid)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toK8sAudit(row)
	return &item, nil
}

func (r *edgeAuditRepo) ListK8sAudits(ctx context.Context, namespaceUID, agentUID string, page, pageSize int) ([]bo.K8sProxyAudit, int, error) {
	query := r.client.K8sProxyAudit.Query().Where(k8sproxyaudit.NamespaceUIDEQ(namespaceUID))
	if agentUID != "" {
		query = query.Where(k8sproxyaudit.AgentUIDEQ(agentUID))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(k8sproxyaudit.FieldCreatedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.K8sProxyAudit, 0, len(rows))
	for _, row := range rows {
		items = append(items, toK8sAudit(row))
	}
	return items, total, nil
}

func toSFTPAudit(row *ent.SFTPAudit) bo.SFTPAudit {
	return bo.SFTPAudit{
		UID: row.UID, NamespaceUID: row.NamespaceUID, AgentUID: row.AgentUID, OperatorUID: row.OperatorUID,
		Path: row.Path, Action: row.Action, Result: row.Result, Detail: row.Detail, CreatedAt: row.CreatedAt,
	}
}

func toK8sAudit(row *ent.K8sProxyAudit) bo.K8sProxyAudit {
	return bo.K8sProxyAudit{
		UID: row.UID, NamespaceUID: row.NamespaceUID, AgentUID: row.AgentUID, OperatorUID: row.OperatorUID,
		Method: row.Method, Resource: row.Resource, KubeNamespace: row.KubeNamespace, Result: row.Result,
		Detail: row.Detail, CreatedAt: row.CreatedAt,
	}
}
