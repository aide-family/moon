package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/k8sproxy"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

type K8sProxyService struct {
	v1.UnimplementedK8SProxyServiceServer
	proxy *k8sproxy.Usecase
}

func NewK8sProxyService(uc *k8sproxy.Usecase) *K8sProxyService {
	return &K8sProxyService{proxy: uc}
}

func (s *K8sProxyService) ProxyK8S(ctx context.Context, req *v1.ProxyK8SRequest) (*v1.ProxyK8SResponse, error) {
	result, err := s.proxy.Proxy(ctx, req.GetAgentUid(), req.GetMethod(), req.GetPath(), req.GetQuery(), req.GetHeadersJson(), req.GetBody())
	if err != nil {
		return nil, err
	}
	return &v1.ProxyK8SResponse{
		StatusCode: result.StatusCode, HeadersJson: result.HeadersJSON, Body: result.Body, Audit: k8sAudit(result.Audit),
	}, nil
}

func (s *K8sProxyService) ListK8SProxyAudits(ctx context.Context, req *v1.ListK8SProxyAuditsRequest) (*v1.ListK8SProxyAuditsResponse, error) {
	items, total, page, pageSize, err := s.proxy.ListAudits(ctx, req.GetAgentUid(), req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.K8SProxyAuditItem, 0, len(items))
	for i := range items {
		out = append(out, k8sAudit(&items[i]))
	}
	return &v1.ListK8SProxyAuditsResponse{Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func k8sAudit(item *bo.K8sProxyAudit) *v1.K8SProxyAuditItem {
	if item == nil {
		return nil
	}
	created := ""
	if !item.CreatedAt.IsZero() {
		created = item.CreatedAt.UTC().Format(time.RFC3339)
	}
	return &v1.K8SProxyAuditItem{
		Uid: item.UID, AgentUid: item.AgentUID, OperatorUid: item.OperatorUID, Method: item.Method,
		Resource: item.Resource, KubeNamespace: item.KubeNamespace, Result: item.Result, Detail: item.Detail, CreatedAt: created,
	}
}
