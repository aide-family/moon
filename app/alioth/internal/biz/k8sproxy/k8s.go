package k8sproxy

import (
	"context"
	"strings"
	"time"

	"github.com/google/wire"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/agenthub"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewUsecase)

type Usecase struct {
	audits repository.EdgeAuditRepo
	agents repository.AgentRepo
	hub    *agenthub.Hub
	access access.Checker
}

type ProxyResult struct {
	StatusCode  int32
	HeadersJSON string
	Body        []byte
	Audit       *bo.K8sProxyAudit
}

func NewUsecase(audits repository.EdgeAuditRepo, agents repository.AgentRepo, hub *agenthub.Hub, checker access.Checker) *Usecase {
	return &Usecase{audits: audits, agents: agents, hub: hub, access: checker}
}

func (uc *Usecase) Proxy(ctx context.Context, agentUID, method, path, query, headersJSON string, body []byte) (*ProxyResult, error) {
	ns, err := uc.access.Allow(ctx, perm.K8sProxy)
	if err != nil {
		return nil, err
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	path = strings.TrimSpace(path)
	if !k8sMethodAllowed(method) || !validK8sPath(path) {
		return nil, merr.Client("invalid method or path")
	}
	agent, err := uc.agents.GetAgent(ctx, ns, agentUID)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if agent.Status != "ACTIVE" {
		return nil, merr.Forbidden("agent is revoked")
	}
	kubeNS := extractKubeNamespace(path)
	if !uc.hub.Online(agentUID) {
		audit, _ := uc.audits.CreateK8sAudit(ctx, bo.K8sProxyAudit{
			NamespaceUID: ns, AgentUID: agentUID, OperatorUID: operatorUID(ctx),
			Method: method, Resource: path, KubeNamespace: kubeNS, Result: "FAILED", Detail: "agent is offline",
		})
		return &ProxyResult{Audit: audit}, merr.Invalid("agent is offline")
	}
	audit, err := uc.audits.CreateK8sAudit(ctx, bo.K8sProxyAudit{
		NamespaceUID: ns, AgentUID: agentUID, OperatorUID: operatorUID(ctx),
		Method: method, Resource: path, KubeNamespace: kubeNS, Result: "PENDING",
	})
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	reqUID := id.New()
	resultCh, err := uc.hub.SendK8s(agentUID, &v1.AgentFrame{K8SRequest: &v1.AgentK8SRequest{
		RequestUid: reqUID, Method: method, Path: path, Query: query, HeadersJson: headersJSON, Body: body,
	}})
	if err != nil {
		updated, _ := uc.audits.UpdateK8sAudit(ctx, audit.UID, "FAILED", err.Error())
		return &ProxyResult{Audit: updated}, err
	}
	timer := time.NewTimer(60 * time.Second)
	defer timer.Stop()
	var result *v1.AgentK8SResult
	select {
	case <-ctx.Done():
		uc.hub.Drop(agentUID, reqUID)
		updated, _ := uc.audits.UpdateK8sAudit(context.WithoutCancel(ctx), audit.UID, "FAILED", "cancelled")
		return &ProxyResult{Audit: updated}, merr.Invalid("cancelled")
	case <-timer.C:
		uc.hub.Drop(agentUID, reqUID)
		updated, _ := uc.audits.UpdateK8sAudit(context.WithoutCancel(ctx), audit.UID, "FAILED", "timed out")
		return &ProxyResult{Audit: updated}, merr.Invalid("timed out")
	case result = <-resultCh:
	}
	if result == nil {
		updated, _ := uc.audits.UpdateK8sAudit(ctx, audit.UID, "FAILED", "agent is offline")
		return &ProxyResult{Audit: updated}, merr.Invalid("agent is offline")
	}
	if result.GetError() != "" {
		updated, _ := uc.audits.UpdateK8sAudit(ctx, audit.UID, "FAILED", result.GetError())
		return &ProxyResult{StatusCode: result.GetStatusCode(), HeadersJSON: result.GetHeadersJson(), Body: result.GetBody(), Audit: updated}, merr.Invalid(result.GetError())
	}
	updated, err := uc.audits.UpdateK8sAudit(ctx, audit.UID, "OK", "")
	if err != nil {
		return &ProxyResult{Audit: audit}, uc.access.Wrap(err)
	}
	return &ProxyResult{
		StatusCode: result.GetStatusCode(), HeadersJSON: result.GetHeadersJson(), Body: result.GetBody(), Audit: updated,
	}, nil
}

func (uc *Usecase) ListAudits(ctx context.Context, agentUID string, pageNo, pageSize int32) ([]bo.K8sProxyAudit, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, "k8s:read")
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.audits.ListK8sAudits(ctx, ns, agentUID, p, s)
	return items, total, p, s, uc.access.Wrap(err)
}

func extractKubeNamespace(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == "namespaces" {
			return parts[i+1]
		}
	}
	return ""
}

// k8sMethodAllowed 只接受 API Server 常用的读写方法，避免 CONNECT/TRACE 穿透代理。
func k8sMethodAllowed(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD":
		return true
	default:
		return false
	}
}

// validK8sPath 只接受以单个 / 开头的 API 路径。拒绝 // 和 @，防止把请求解析到别的主机。
func validK8sPath(path string) bool {
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.ContainsAny(path, "\\@") {
		return false
	}
	for _, segment := range strings.Split(path, "/") {
		if segment == ".." {
			return false
		}
	}
	return true
}

func operatorUID(ctx context.Context) string {
	uid, _ := actor.UserID(ctx)
	return uid
}
