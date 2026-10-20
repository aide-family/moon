package ssh

import (
	"context"
	"strings"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"
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

const maxOutputBytes = 64 * 1024

var ProviderSet = wire.NewSet(NewUsecase)

// Usecase 负责 SSH 命令审核与执行。
type Usecase struct {
	commands repository.SSHCommandRepo
	agents   repository.AgentRepo
	hub      *agenthub.Hub
	access   access.Checker
}

func NewUsecase(commands repository.SSHCommandRepo, agents repository.AgentRepo, hub *agenthub.Hub, checker access.Checker) *Usecase {
	return &Usecase{commands: commands, agents: agents, hub: hub, access: checker}
}

func (uc *Usecase) Create(ctx context.Context, in bo.SSHCommand) (*bo.SSHCommand, error) {
	ns, err := uc.access.Allow(ctx, "ssh:write")
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	in.Command = strings.TrimSpace(in.Command)
	if in.Name == "" || in.Command == "" {
		return nil, merr.Client("name and command are required")
	}
	item, err := uc.commands.CreateCommand(ctx, in, operatorUID(ctx))
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) Update(ctx context.Context, in bo.SSHCommand) (*bo.SSHCommand, error) {
	ns, err := uc.access.Allow(ctx, "ssh:write")
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	in.Command = strings.TrimSpace(in.Command)
	if in.Name == "" || in.Command == "" {
		return nil, merr.Client("name and command are required")
	}
	item, err := uc.commands.UpdateCommandPending(ctx, in, operatorUID(ctx))
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) Approve(ctx context.Context, uid string) (*bo.SSHCommand, error) {
	ns, err := uc.access.Allow(ctx, "ssh:write")
	if err != nil {
		return nil, err
	}
	item, err := uc.commands.ApproveCommand(ctx, ns, uid, operatorUID(ctx))
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) Reject(ctx context.Context, uid string) (*bo.SSHCommand, error) {
	ns, err := uc.access.Allow(ctx, "ssh:write")
	if err != nil {
		return nil, err
	}
	item, err := uc.commands.RejectCommand(ctx, ns, uid, operatorUID(ctx))
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) Delete(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, "ssh:write")
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.commands.DeleteCommand(ctx, ns, uid))
}

func (uc *Usecase) Get(ctx context.Context, uid string) (*bo.SSHCommand, error) {
	ns, err := uc.access.Allow(ctx, "ssh:read")
	if err != nil {
		return nil, err
	}
	item, err := uc.commands.GetCommand(ctx, ns, uid)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) List(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.SSHCommand, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, "ssh:read")
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.commands.ListCommands(ctx, ns, p, s, strings.TrimSpace(name))
	return items, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) ListAudits(ctx context.Context, commandUID, agentUID string, pageNo, pageSize int32) ([]bo.SSHCommandAudit, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, "ssh:read")
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.commands.ListAudits(ctx, ns, commandUID, agentUID, p, s)
	return items, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) Execute(ctx context.Context, commandUID, agentUID string) (*bo.SSHCommandAudit, error) {
	ns, err := uc.access.Allow(ctx, "ssh:exec")
	if err != nil {
		return nil, err
	}
	cmd, err := uc.commands.GetCommand(ctx, ns, commandUID)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if cmd.ReviewStatus != "APPROVED" {
		return nil, merr.Forbidden("command is not approved")
	}
	machine, err := uc.agents.GetAgent(ctx, ns, agentUID)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if machine.Status != "ACTIVE" {
		uc.recordExec(ctx, ns, cmd, agentUID, "FAILED", "agent is revoked", 0, "", false)
		return nil, merr.Forbidden("agent is revoked")
	}
	if !uc.hub.Online(agentUID) {
		uc.recordExec(ctx, ns, cmd, agentUID, "FAILED", "agent is offline", 0, "", false)
		return nil, merr.Invalid("agent is offline")
	}
	password, privateKey, err := uc.agents.GetSSHSecrets(ctx, agentUID)
	if err != nil {
		uc.recordExec(ctx, ns, cmd, agentUID, "FAILED", err.Error(), 0, "", false)
		return nil, uc.access.Wrap(err)
	}
	requestUID := id.New()
	resultCh, err := uc.hub.SendExec(agentUID, &v1.AgentFrame{
		ExecRequest: &v1.AgentExecRequest{
			RequestUid:    requestUID,
			Command:       cmd.Command,
			Workdir:       cmd.Workdir,
			Env:           cmd.Env,
			SshHost:       machine.SSHHost,
			SshPort:       int32(machine.SSHPort),
			SshUser:       machine.SSHUser,
			SshPassword:   password,
			SshPrivateKey: privateKey,
		},
	})
	if err != nil {
		uc.recordExec(ctx, ns, cmd, agentUID, "FAILED", err.Error(), 0, "", false)
		return nil, err
	}
	timer := time.NewTimer(60 * time.Second)
	defer timer.Stop()
	var result *v1.AgentExecResult
	select {
	case <-ctx.Done():
		uc.hub.Drop(agentUID, requestUID)
		uc.recordExec(ctx, ns, cmd, agentUID, "FAILED", "execute cancelled", 0, "", false)
		return nil, merr.Invalid("execute cancelled")
	case <-timer.C:
		uc.hub.Drop(agentUID, requestUID)
		uc.recordExec(ctx, ns, cmd, agentUID, "FAILED", "execute timed out", 0, "", false)
		return nil, merr.Invalid("execute timed out")
	case result = <-resultCh:
	}
	if result == nil {
		uc.recordExec(ctx, ns, cmd, agentUID, "FAILED", "agent is offline", 0, "", false)
		return nil, merr.Invalid("agent is offline")
	}
	output := result.GetOutput()
	truncated := result.GetTruncated()
	if len(output) > maxOutputBytes {
		output = output[:maxOutputBytes]
		truncated = true
	}
	status := "DONE"
	detail := result.GetError()
	if detail != "" || result.GetExitCode() != 0 {
		status = "FAILED"
	}
	audit, err := uc.commands.AddExecAudit(ctx, bo.SSHCommandAudit{
		NamespaceUID: ns,
		CommandUID:   commandUID,
		AgentUID:     agentUID,
		OperatorUID:  operatorUID(ctx),
		Status:       status,
		Command:      cmd.Command,
		ExitCode:     int(result.GetExitCode()),
		Output:       output,
		Truncated:    truncated,
		Detail:       detail,
	})
	return audit, uc.access.Wrap(err)
}

// ExecuteOnAgents 把同一条已通过的命令分别送到多台机器。一台失败仍继续其余机器，并为失败的机器留下审计。
func (uc *Usecase) ExecuteOnAgents(ctx context.Context, commandUID string, agentUIDs []string) ([]bo.SSHCommandAudit, error) {
	if len(agentUIDs) == 0 {
		return nil, merr.Invalid("select an online agent")
	}
	audits := make([]bo.SSHCommandAudit, 0, len(agentUIDs))
	for _, agentUID := range agentUIDs {
		audit, err := uc.Execute(ctx, commandUID, agentUID)
		if err != nil {
			if abortAgentBatch(err) {
				return nil, err
			}
			failed, ferr := uc.failExec(ctx, commandUID, agentUID, err)
			if ferr != nil {
				return nil, ferr
			}
			audits = append(audits, *failed)
			continue
		}
		audits = append(audits, *audit)
	}
	return audits, nil
}

func (uc *Usecase) failExec(ctx context.Context, commandUID, agentUID string, cause error) (*bo.SSHCommandAudit, error) {
	ns, err := uc.access.Allow(ctx, "ssh:exec")
	if err != nil {
		return nil, err
	}
	command := commandUID
	if cmd, err := uc.commands.GetCommand(ctx, ns, commandUID); err == nil {
		command = cmd.Command
	}
	detail := kerrors.FromError(cause).Message
	if detail == "" {
		detail = "execute failed"
	}
	return uc.commands.AddExecAudit(ctx, bo.SSHCommandAudit{
		NamespaceUID: ns,
		CommandUID:   commandUID,
		AgentUID:     agentUID,
		OperatorUID:  operatorUID(ctx),
		Status:       "FAILED",
		Command:      command,
		ExitCode:     1,
		Detail:       detail,
	})
}

func abortAgentBatch(err error) bool {
	e := kerrors.FromError(err)
	if e.Reason == merr.ReasonUnauthenticated {
		return true
	}
	if e.Reason == merr.ReasonForbidden && e.Message != "agent is revoked" {
		return true
	}
	return false
}

func (uc *Usecase) ExecuteMany(ctx context.Context, commandUIDs []string, agentUID string) ([]bo.SSHCommandAudit, error) {
	audits := make([]bo.SSHCommandAudit, 0, len(commandUIDs))
	for _, uid := range commandUIDs {
		audit, err := uc.Execute(ctx, uid, agentUID)
		if err != nil {
			return audits, err
		}
		audits = append(audits, *audit)
	}
	return audits, nil
}

// recordExec 为失败路径补审计。请求可能已超时或取消，因此使用 WithoutCancel 保证审计仍能落库。
func (uc *Usecase) recordExec(ctx context.Context, namespaceUID string, cmd *bo.SSHCommand, agentUID, status, detail string, exitCode int, output string, truncated bool) {
	if cmd == nil {
		return
	}
	_, _ = uc.commands.AddExecAudit(context.WithoutCancel(ctx), bo.SSHCommandAudit{
		NamespaceUID: namespaceUID,
		CommandUID:   cmd.UID,
		AgentUID:     agentUID,
		OperatorUID:  operatorUID(ctx),
		Status:       status,
		Command:      cmd.Command,
		ExitCode:     exitCode,
		Output:       output,
		Truncated:    truncated,
		Detail:       detail,
	})
}

func operatorUID(ctx context.Context) string {
	uid, _ := actor.UserID(ctx)
	return uid
}
