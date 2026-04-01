package biz

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
)

// SSHCommand handles SSH command templates, audits, and remote execution.
type SSHCommand struct {
	commands repository.SSHCommand
	audits   repository.CommandAudit
	ssh      repository.SSHOperator
	helper   *klog.Helper
}

func NewSSHCommand(commands repository.SSHCommand, audits repository.CommandAudit, ssh repository.SSHOperator, helper *klog.Helper) *SSHCommand {
	return &SSHCommand{commands: commands, audits: audits, ssh: ssh, helper: helper}
}

func (u *SSHCommand) requireUser(ctx context.Context) (snowflake.ID, error) {
	uid := contextx.GetUserUID(ctx)
	if uid == 0 {
		return 0, merr.ErrorInvalidArgument("user context is required")
	}
	return uid, nil
}

func (u *SSHCommand) SubmitCreate(ctx context.Context, fields *bo.SSHCommandFields) (*bo.SSHCommandAuditItemBo, error) {
	creator, err := u.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	if fields == nil {
		return nil, merr.ErrorInvalidArgument("command fields are required")
	}
	return u.audits.Create(ctx, &bo.CommandAuditCreateRepoBo{
		Creator:         creator,
		TargetCommandID: 0,
		Kind:            enum.SSHCommandAuditKind_SSHCommandAuditKind_CREATE,
		Fields:          *fields,
	})
}

func (u *SSHCommand) SubmitUpdate(ctx context.Context, in *bo.SubmitSSHCommandUpdateInput) (*bo.SSHCommandAuditItemBo, error) {
	creator, err := u.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, merr.ErrorInvalidArgument("update submission is required")
	}
	if _, err := u.commands.Get(ctx, in.CommandUID); err != nil {
		return nil, err
	}
	return u.audits.Create(ctx, &bo.CommandAuditCreateRepoBo{
		Creator:         creator,
		TargetCommandID: in.CommandUID,
		Kind:            enum.SSHCommandAuditKind_SSHCommandAuditKind_UPDATE,
		Fields:          in.Fields,
	})
}

func (u *SSHCommand) ListCommands(ctx context.Context, req *bo.ListSSHCommandsBo) (*bo.PageResponseBo[*bo.SSHCommandItemBo], error) {
	return u.commands.List(ctx, req)
}

func (u *SSHCommand) GetCommand(ctx context.Context, uid snowflake.ID) (*bo.SSHCommandItemBo, error) {
	return u.commands.Get(ctx, uid)
}

func (u *SSHCommand) ListAudits(ctx context.Context, req *bo.ListSSHCommandAuditsBo) (*bo.PageResponseBo[*bo.SSHCommandAuditItemBo], error) {
	return u.audits.List(ctx, req)
}

func (u *SSHCommand) GetAudit(ctx context.Context, uid snowflake.ID) (*bo.SSHCommandAuditItemBo, error) {
	return u.audits.Get(ctx, uid)
}

func (u *SSHCommand) ApproveAudit(ctx context.Context, uid snowflake.ID) (*bo.SSHCommandAuditItemBo, *bo.SSHCommandItemBo, error) {
	reviewer, err := u.requireUser(ctx)
	if err != nil {
		return nil, nil, err
	}
	return u.audits.Approve(ctx, uid, reviewer)
}

func (u *SSHCommand) RejectAudit(ctx context.Context, in *bo.RejectSSHCommandAuditInput) (*bo.SSHCommandAuditItemBo, error) {
	reviewer, err := u.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, merr.ErrorInvalidArgument("reject input is required")
	}
	return u.audits.Reject(ctx, &bo.CommandAuditRejectBo{
		AuditUID: in.AuditUID,
		Reviewer: reviewer,
		Reason:   in.Reason,
	})
}

func (u *SSHCommand) Execute(ctx context.Context, in *bo.ExecuteStoredSSHCommandBo) (*bo.SSHExecReply, error) {
	if in == nil {
		return nil, merr.ErrorInvalidArgument("execute input is required")
	}
	cmd, err := u.commands.Get(ctx, in.CommandUID)
	if err != nil {
		return nil, err
	}
	if cmd.Disabled {
		return nil, merr.ErrorInvalidArgument("ssh command is disabled")
	}
	timeout := time.Duration(in.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	env := cmd.Env
	if env == nil {
		env = map[string]string{}
	}
	req := &bo.SSHExecRequest{
		Host:       in.Host,
		Port:       in.Port,
		Username:   in.Username,
		Password:   in.Password,
		PrivateKey: in.PrivateKey,
		Timeout:    timeout,
		Command:    cmd.Content,
		WorkDir:    cmd.WorkDir,
		Env:        env,
	}
	return u.ssh.Exec(ctx, req)
}
