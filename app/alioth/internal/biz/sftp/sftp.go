package sftp

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
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewUsecase)

type Usecase struct {
	audits repository.EdgeAuditRepo
	agents repository.AgentRepo
	hub    *agenthub.Hub
	access access.Checker
}

type outcome struct {
	entries []bo.SFTPEntry
	content []byte
	audit   *bo.SFTPAudit
}

func NewUsecase(audits repository.EdgeAuditRepo, agents repository.AgentRepo, hub *agenthub.Hub, checker access.Checker) *Usecase {
	return &Usecase{audits: audits, agents: agents, hub: hub, access: checker}
}

func (uc *Usecase) List(ctx context.Context, agentUID, path string) ([]bo.SFTPEntry, *bo.SFTPAudit, error) {
	out, err := uc.run(ctx, "sftp:read", agentUID, path, "LIST", nil)
	return out.entries, out.audit, err
}

func (uc *Usecase) Download(ctx context.Context, agentUID, path string) ([]byte, *bo.SFTPAudit, error) {
	out, err := uc.run(ctx, "sftp:read", agentUID, path, "DOWNLOAD", nil)
	return out.content, out.audit, err
}

func (uc *Usecase) Upload(ctx context.Context, agentUID, path string, content []byte) (*bo.SFTPAudit, error) {
	out, err := uc.run(ctx, "sftp:write", agentUID, path, "UPLOAD", content)
	return out.audit, err
}

func (uc *Usecase) Mkdir(ctx context.Context, agentUID, path string) (*bo.SFTPAudit, error) {
	out, err := uc.run(ctx, "sftp:write", agentUID, path, "MKDIR", nil)
	return out.audit, err
}

func (uc *Usecase) Delete(ctx context.Context, agentUID, path string) (*bo.SFTPAudit, error) {
	out, err := uc.run(ctx, "sftp:write", agentUID, path, "DELETE", nil)
	return out.audit, err
}

func (uc *Usecase) ListAudits(ctx context.Context, agentUID string, pageNo, pageSize int32) ([]bo.SFTPAudit, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, "sftp:read")
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.audits.ListSFTPAudits(ctx, ns, agentUID, p, s)
	return items, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) run(ctx context.Context, perm, agentUID, path, action string, content []byte) (outcome, error) {
	ns, err := uc.access.Allow(ctx, perm)
	if err != nil {
		return outcome{}, err
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return outcome{}, merr.Client("path is required")
	}
	machine, err := uc.agents.GetAgent(ctx, ns, agentUID)
	if err != nil {
		return outcome{}, uc.access.Wrap(err)
	}
	if !uc.hub.Online(agentUID) {
		audit, _ := uc.audits.CreateSFTPAudit(ctx, bo.SFTPAudit{
			NamespaceUID: ns, AgentUID: agentUID, OperatorUID: operatorUID(ctx),
			Path: path, Action: action, Result: "FAILED", Detail: "agent is offline",
		})
		return outcome{audit: audit}, merr.Invalid("agent is offline")
	}
	audit, err := uc.audits.CreateSFTPAudit(ctx, bo.SFTPAudit{
		NamespaceUID: ns, AgentUID: agentUID, OperatorUID: operatorUID(ctx),
		Path: path, Action: action, Result: "PENDING",
	})
	if err != nil {
		return outcome{}, uc.access.Wrap(err)
	}
	password, key, err := uc.agents.GetSSHSecrets(ctx, agentUID)
	if err != nil {
		updated, _ := uc.audits.UpdateSFTPAudit(ctx, audit.UID, "FAILED", err.Error())
		return outcome{audit: updated}, uc.access.Wrap(err)
	}
	reqUID := id.New()
	resultCh, err := uc.hub.SendSFTP(agentUID, &v1.AgentFrame{SftpRequest: &v1.AgentSFTPRequest{
		RequestUid: reqUID, Action: action, Path: path, Content: content,
		SshHost: machine.SSHHost, SshPort: int32(machine.SSHPort), SshUser: machine.SSHUser,
		SshPassword: password, SshPrivateKey: key,
	}})
	if err != nil {
		updated, _ := uc.audits.UpdateSFTPAudit(ctx, audit.UID, "FAILED", err.Error())
		return outcome{audit: updated}, err
	}
	timer := time.NewTimer(60 * time.Second)
	defer timer.Stop()
	var result *v1.AgentSFTPResult
	select {
	case <-ctx.Done():
		updated, _ := uc.audits.UpdateSFTPAudit(ctx, audit.UID, "FAILED", "cancelled")
		return outcome{audit: updated}, merr.Invalid("cancelled")
	case <-timer.C:
		updated, _ := uc.audits.UpdateSFTPAudit(ctx, audit.UID, "FAILED", "timed out")
		return outcome{audit: updated}, merr.Invalid("timed out")
	case result = <-resultCh:
	}
	if result == nil || !result.GetOk() {
		detail := "agent is offline"
		if result != nil && result.GetError() != "" {
			detail = result.GetError()
		}
		updated, _ := uc.audits.UpdateSFTPAudit(ctx, audit.UID, "FAILED", detail)
		return outcome{audit: updated}, merr.Invalid(detail)
	}
	updated, err := uc.audits.UpdateSFTPAudit(ctx, audit.UID, "OK", "")
	if err != nil {
		return outcome{audit: audit}, uc.access.Wrap(err)
	}
	entries := make([]bo.SFTPEntry, 0, len(result.GetEntries()))
	for _, item := range result.GetEntries() {
		entries = append(entries, bo.SFTPEntry{
			Name: item.GetName(), Path: item.GetPath(), IsDir: item.GetIsDir(),
			Size: item.GetSize(), Mode: item.GetMode(), ModTime: item.GetModTime(),
		})
	}
	return outcome{entries: entries, content: result.GetContent(), audit: updated}, nil
}

func operatorUID(ctx context.Context) string {
	uid, _ := actor.UserID(ctx)
	return uid
}
