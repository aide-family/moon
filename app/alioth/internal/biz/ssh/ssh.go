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

var ProviderSet = wire.NewSet(NewExecutor)

// Executor 负责 SSH 命令审核与执行。
type Executor struct {
	commands repository.SSHCommandRepo
	agents   repository.AgentRepo
	hub      *agenthub.Hub
	access   access.Checker
}

func NewExecutor(commands repository.SSHCommandRepo, agents repository.AgentRepo, hub *agenthub.Hub, checker access.Checker) *Executor {
	return &Executor{commands: commands, agents: agents, hub: hub, access: checker}
}

func (executor *Executor) Create(ctx context.Context, in bo.SSHCommand) (*bo.SSHCommand, error) {
	ns, err := executor.access.Allow(ctx, "ssh:write")
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	in.Command = strings.TrimSpace(in.Command)
	if in.Name == "" || in.Command == "" {
		return nil, merr.Client("name and command are required")
	}
	item, err := executor.commands.CreateCommand(ctx, in, operatorUID(ctx))
	return item, executor.access.Wrap(err)
}

func (executor *Executor) Update(ctx context.Context, in bo.SSHCommand) (*bo.SSHCommand, error) {
	ns, err := executor.access.Allow(ctx, "ssh:write")
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	in.Command = strings.TrimSpace(in.Command)
	if in.Name == "" || in.Command == "" {
		return nil, merr.Client("name and command are required")
	}
	item, err := executor.commands.UpdateCommandPending(ctx, in, operatorUID(ctx))
	return item, executor.access.Wrap(err)
}

func (executor *Executor) Approve(ctx context.Context, uid string) (*bo.SSHCommand, error) {
	ns, err := executor.access.Allow(ctx, "ssh:write")
	if err != nil {
		return nil, err
	}
	item, err := executor.commands.ApproveCommand(ctx, ns, uid, operatorUID(ctx))
	return item, executor.access.Wrap(err)
}

func (executor *Executor) Reject(ctx context.Context, uid string) (*bo.SSHCommand, error) {
	ns, err := executor.access.Allow(ctx, "ssh:write")
	if err != nil {
		return nil, err
	}
	item, err := executor.commands.RejectCommand(ctx, ns, uid, operatorUID(ctx))
	return item, executor.access.Wrap(err)
}

func (executor *Executor) Delete(ctx context.Context, uid string) error {
	ns, err := executor.access.Allow(ctx, "ssh:write")
	if err != nil {
		return err
	}
	return executor.access.Wrap(executor.commands.DeleteCommand(ctx, ns, uid))
}

func (executor *Executor) Get(ctx context.Context, uid string) (*bo.SSHCommand, error) {
	ns, err := executor.access.Allow(ctx, "ssh:read")
	if err != nil {
		return nil, err
	}
	item, err := executor.commands.GetCommand(ctx, ns, uid)
	return item, executor.access.Wrap(err)
}

func (executor *Executor) List(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.SSHCommand, int, int, int, error) {
	ns, err := executor.access.Allow(ctx, "ssh:read")
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := executor.commands.ListCommands(ctx, ns, p, s, strings.TrimSpace(name))
	return items, total, p, s, executor.access.Wrap(err)
}

func (executor *Executor) ListAudits(ctx context.Context, commandUID, agentUID string, pageNo, pageSize int32) ([]bo.SSHCommandAudit, int, int, int, error) {
	ns, err := executor.access.Allow(ctx, "ssh:read")
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := executor.commands.ListAudits(ctx, ns, commandUID, agentUID, p, s)
	return items, total, p, s, executor.access.Wrap(err)
}

func (executor *Executor) Execute(ctx context.Context, commandUID, agentUID string) (*bo.SSHCommandAudit, error) {
	ns, err := executor.access.Allow(ctx, "ssh:exec")
	if err != nil {
		return nil, err
	}
	cmd, err := executor.commands.GetCommand(ctx, ns, commandUID)
	if err != nil {
		return nil, executor.access.Wrap(err)
	}
	if cmd.ReviewStatus != "APPROVED" {
		return nil, merr.Forbidden("command is not approved")
	}
	machine, err := executor.agents.GetAgent(ctx, ns, agentUID)
	if err != nil {
		return nil, executor.access.Wrap(err)
	}
	if machine.Status != "ACTIVE" {
		executor.recordExec(ctx, ns, cmd, agentUID, "FAILED", "agent is revoked", 0, "", false)
		return nil, merr.Forbidden("agent is revoked")
	}
	if !executor.hub.Online(agentUID) {
		executor.recordExec(ctx, ns, cmd, agentUID, "FAILED", "agent is offline", 0, "", false)
		return nil, merr.Invalid("agent is offline")
	}
	requestUID := id.New()
	resultCh, err := executor.hub.SendExec(agentUID, &v1.AgentFrame{
		ExecRequest: &v1.AgentExecRequest{
			RequestUid: requestUID,
			Command:    cmd.Command,
			Workdir:    cmd.Workdir,
			Env:        cmd.Env,
		},
	})
	if err != nil {
		executor.recordExec(ctx, ns, cmd, agentUID, "FAILED", err.Error(), 0, "", false)
		return nil, err
	}
	timer := time.NewTimer(60 * time.Second)
	defer timer.Stop()
	var result *v1.AgentExecResult
	select {
	case <-ctx.Done():
		executor.hub.Drop(agentUID, requestUID)
		executor.recordExec(ctx, ns, cmd, agentUID, "FAILED", "execute cancelled", 0, "", false)
		return nil, merr.Invalid("execute cancelled")
	case <-timer.C:
		executor.hub.Drop(agentUID, requestUID)
		executor.recordExec(ctx, ns, cmd, agentUID, "FAILED", "execute timed out", 0, "", false)
		return nil, merr.Invalid("execute timed out")
	case result = <-resultCh:
	}
	if result == nil {
		executor.recordExec(ctx, ns, cmd, agentUID, "FAILED", "agent is offline", 0, "", false)
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
	audit, err := executor.commands.AddExecAudit(ctx, bo.SSHCommandAudit{
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
	return audit, executor.access.Wrap(err)
}

// ExecuteOnAgents 把同一条已通过的命令分别送到多台机器。一台失败仍继续其余机器，并为失败的机器留下审计。
func (executor *Executor) ExecuteOnAgents(ctx context.Context, commandUID string, agentUIDs []string) ([]bo.SSHCommandAudit, error) {
	if len(agentUIDs) == 0 {
		return nil, merr.Invalid("select an online agent")
	}
	audits := make([]bo.SSHCommandAudit, 0, len(agentUIDs))
	for _, agentUID := range agentUIDs {
		audit, err := executor.Execute(ctx, commandUID, agentUID)
		if err != nil {
			if abortAgentBatch(err) {
				return nil, err
			}
			failed, ferr := executor.failExec(ctx, commandUID, agentUID, err)
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

func (executor *Executor) failExec(ctx context.Context, commandUID, agentUID string, cause error) (*bo.SSHCommandAudit, error) {
	ns, err := executor.access.Allow(ctx, "ssh:exec")
	if err != nil {
		return nil, err
	}
	command := commandUID
	if cmd, err := executor.commands.GetCommand(ctx, ns, commandUID); err == nil {
		command = cmd.Command
	}
	detail := kerrors.FromError(cause).Message
	if detail == "" {
		detail = "execute failed"
	}
	return executor.commands.AddExecAudit(ctx, bo.SSHCommandAudit{
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

func (executor *Executor) ExecuteMany(ctx context.Context, commandUIDs []string, agentUID string) ([]bo.SSHCommandAudit, error) {
	audits := make([]bo.SSHCommandAudit, 0, len(commandUIDs))
	for _, uid := range commandUIDs {
		audit, err := executor.Execute(ctx, uid, agentUID)
		if err != nil {
			return audits, err
		}
		audits = append(audits, *audit)
	}
	return audits, nil
}

// recordExec 为失败路径补审计。请求可能已超时或取消，因此使用 WithoutCancel 保证审计仍能落库。
func (executor *Executor) recordExec(ctx context.Context, namespaceUID string, cmd *bo.SSHCommand, agentUID, status, detail string, exitCode int, output string, truncated bool) {
	if cmd == nil {
		return
	}
	_, _ = executor.commands.AddExecAudit(context.WithoutCancel(ctx), bo.SSHCommandAudit{
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
