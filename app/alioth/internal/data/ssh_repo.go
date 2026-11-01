package data

import (
	"context"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/agent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/sshcommand"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/sshcommandaudit"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *sshCommandRepo) CreateCommand(ctx context.Context, in bo.SSHCommand, operatorUID string) (*bo.SSHCommand, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.SSHCommand.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetCommand(in.Command).
		SetWorkdir(in.Workdir).
		SetEnv(in.Env).
		SetReviewStatus(sshcommand.ReviewStatusPENDING).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := tx.SSHCommandAudit.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetCommandUID(row.UID).
		SetOperatorUID(operatorUID).
		SetKind(sshcommandaudit.KindCREATE).
		SetStatus(sshcommandaudit.StatusPENDING).
		SetCommand(in.Command).
		Exec(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	item := toSSHCommand(row)
	return &item, nil
}

func (r *sshCommandRepo) UpdateCommandPending(ctx context.Context, in bo.SSHCommand, operatorUID string) (*bo.SSHCommand, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	n, err := tx.SSHCommand.Update().Where(
		sshcommand.NamespaceUIDEQ(in.NamespaceUID), sshcommand.UIDEQ(in.UID), sshcommand.DeletedAtIsNil(),
	).
		SetName(in.Name).
		SetCommand(in.Command).
		SetWorkdir(in.Workdir).
		SetEnv(in.Env).
		SetRemark(in.Remark).
		SetReviewStatus(sshcommand.ReviewStatusPENDING).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	if err := tx.SSHCommandAudit.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetCommandUID(in.UID).
		SetOperatorUID(operatorUID).
		SetKind(sshcommandaudit.KindUPDATE).
		SetStatus(sshcommandaudit.StatusPENDING).
		SetCommand(in.Command).
		Exec(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetCommand(ctx, in.NamespaceUID, in.UID)
}

func (r *sshCommandRepo) ApproveCommand(ctx context.Context, namespaceUID, uid, operatorUID string) (*bo.SSHCommand, error) {
	return r.setReview(ctx, namespaceUID, uid, operatorUID, sshcommand.ReviewStatusAPPROVED, sshcommandaudit.KindAPPROVE, sshcommandaudit.StatusAPPROVED)
}

func (r *sshCommandRepo) RejectCommand(ctx context.Context, namespaceUID, uid, operatorUID string) (*bo.SSHCommand, error) {
	return r.setReview(ctx, namespaceUID, uid, operatorUID, sshcommand.ReviewStatusREJECTED, sshcommandaudit.KindREJECT, sshcommandaudit.StatusREJECTED)
}

func (r *sshCommandRepo) setReview(ctx context.Context, namespaceUID, uid, operatorUID string, review sshcommand.ReviewStatus, kind sshcommandaudit.Kind, status sshcommandaudit.Status) (*bo.SSHCommand, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.SSHCommand.Query().Where(sshcommand.NamespaceUIDEQ(namespaceUID), sshcommand.UIDEQ(uid), sshcommand.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	updated, err := tx.SSHCommand.UpdateOne(row).SetReviewStatus(review).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := tx.SSHCommandAudit.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespaceUID).
		SetCommandUID(uid).
		SetOperatorUID(operatorUID).
		SetKind(kind).
		SetStatus(status).
		SetCommand(updated.Command).
		Exec(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	item := toSSHCommand(updated)
	return &item, nil
}

func (r *sshCommandRepo) DeleteCommand(ctx context.Context, namespaceUID, uid string) error {
	n, err := r.client.SSHCommand.Update().Where(sshcommand.NamespaceUIDEQ(namespaceUID), sshcommand.UIDEQ(uid), sshcommand.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

func (r *sshCommandRepo) GetCommand(ctx context.Context, namespaceUID, uid string) (*bo.SSHCommand, error) {
	row, err := r.client.SSHCommand.Query().Where(sshcommand.NamespaceUIDEQ(namespaceUID), sshcommand.UIDEQ(uid), sshcommand.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toSSHCommand(row)
	return &item, nil
}

func (r *sshCommandRepo) ListCommands(ctx context.Context, namespaceUID string, page, pageSize int, name, reviewStatus string) ([]bo.SSHCommand, int, error) {
	query := r.client.SSHCommand.Query().Where(sshcommand.NamespaceUIDEQ(namespaceUID), sshcommand.DeletedAtIsNil())
	if name != "" {
		query = query.Where(sshcommand.NameContainsFold(name))
	}
	if reviewStatus != "" {
		query = query.Where(sshcommand.ReviewStatusEQ(sshcommand.ReviewStatus(reviewStatus)))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(sshcommand.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.SSHCommand, 0, len(rows))
	for _, row := range rows {
		items = append(items, toSSHCommand(row))
	}
	return items, total, nil
}

func (r *sshCommandRepo) AddExecAudit(ctx context.Context, in bo.SSHCommandAudit) (*bo.SSHCommandAudit, error) {
	row, err := r.client.SSHCommandAudit.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetCommandUID(in.CommandUID).
		SetAgentUID(in.AgentUID).
		SetOperatorUID(in.OperatorUID).
		SetKind(sshcommandaudit.KindEXECUTE).
		SetStatus(sshcommandaudit.Status(in.Status)).
		SetCommand(in.Command).
		SetExitCode(in.ExitCode).
		SetOutput(in.Output).
		SetTruncated(in.Truncated).
		SetDetail(in.Detail).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toSSHAudit(row)
	return &item, nil
}

func (r *sshCommandRepo) ListAudits(ctx context.Context, namespaceUID string, filter bo.SSHCommandAuditQuery, page, pageSize int) ([]bo.SSHCommandAudit, int, error) {
	auditQuery := r.client.SSHCommandAudit.Query().Where(sshcommandaudit.NamespaceUIDEQ(namespaceUID))
	if filter.CommandUID != "" {
		auditQuery = auditQuery.Where(sshcommandaudit.CommandUIDEQ(filter.CommandUID))
	}
	if filter.AgentUID != "" {
		auditQuery = auditQuery.Where(sshcommandaudit.AgentUIDEQ(filter.AgentUID))
	}
	if filter.Kind != "" {
		auditQuery = auditQuery.Where(sshcommandaudit.KindEQ(sshcommandaudit.Kind(filter.Kind)))
	}
	if filter.Status != "" {
		auditQuery = auditQuery.Where(sshcommandaudit.StatusEQ(sshcommandaudit.Status(filter.Status)))
	}
	if filter.OperatorUID != "" {
		auditQuery = auditQuery.Where(sshcommandaudit.OperatorUIDEQ(filter.OperatorUID))
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		auditQuery = auditQuery.Where(sshcommandaudit.CommandContainsFold(keyword))
	}
	if filter.StartTime > 0 {
		auditQuery = auditQuery.Where(sshcommandaudit.CreatedAtGTE(time.Unix(filter.StartTime, 0)))
	}
	if filter.EndTime > 0 {
		auditQuery = auditQuery.Where(sshcommandaudit.CreatedAtLT(time.Unix(filter.EndTime, 0)))
	}
	total, err := auditQuery.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := auditQuery.Order(ent.Desc(sshcommandaudit.FieldCreatedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.SSHCommandAudit, 0, len(rows))
	operatorUIDs := make([]string, 0, len(rows))
	seenOperators := make(map[string]struct{}, len(rows))
	agentUIDs := make([]string, 0, len(rows))
	seenAgents := make(map[string]struct{}, len(rows))
	commandUIDs := make([]string, 0, len(rows))
	seenCommands := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		item := toSSHAudit(row)
		items = append(items, item)
		if item.OperatorUID != "" {
			if _, ok := seenOperators[item.OperatorUID]; !ok {
				seenOperators[item.OperatorUID] = struct{}{}
				operatorUIDs = append(operatorUIDs, item.OperatorUID)
			}
		}
		if item.AgentUID != "" {
			if _, ok := seenAgents[item.AgentUID]; !ok {
				seenAgents[item.AgentUID] = struct{}{}
				agentUIDs = append(agentUIDs, item.AgentUID)
			}
		}
		if item.CommandUID != "" {
			if _, ok := seenCommands[item.CommandUID]; !ok {
				seenCommands[item.CommandUID] = struct{}{}
				commandUIDs = append(commandUIDs, item.CommandUID)
			}
		}
	}
	if len(operatorUIDs) > 0 {
		users, err := r.client.User.Query().Where(user.UIDIn(operatorUIDs...)).All(ctx)
		if err != nil {
			return nil, 0, r.fail(err)
		}
		profiles := make(map[string]*ent.User, len(users))
		for _, account := range users {
			profiles[account.UID] = account
		}
		for i := range items {
			if account, ok := profiles[items[i].OperatorUID]; ok {
				items[i].OperatorName = account.Name
				items[i].OperatorNickname = account.Nickname
				items[i].OperatorAvatar = account.Avatar
			}
		}
	}
	// 机器名与命令名按 uid 批量补全。审计已经按租户过滤，uid 全局唯一，因此这里不再加租户条件；
	// 已被删除的机器或命令不再返回名字，join 失败也不能丢审计行。
	if len(agentUIDs) > 0 {
		agents, err := r.client.Agent.Query().Where(agent.UIDIn(agentUIDs...), agent.DeletedAtIsNil()).All(ctx)
		if err != nil {
			return nil, 0, r.fail(err)
		}
		names := make(map[string]string, len(agents))
		for _, machine := range agents {
			names[machine.UID] = machine.Name
		}
		for i := range items {
			items[i].AgentName = names[items[i].AgentUID]
		}
	}
	if len(commandUIDs) > 0 {
		commands, err := r.client.SSHCommand.Query().Where(sshcommand.UIDIn(commandUIDs...), sshcommand.DeletedAtIsNil()).All(ctx)
		if err != nil {
			return nil, 0, r.fail(err)
		}
		names := make(map[string]string, len(commands))
		for _, command := range commands {
			names[command.UID] = command.Name
		}
		for i := range items {
			items[i].CommandName = names[items[i].CommandUID]
		}
	}
	return items, total, nil
}

func toSSHCommand(row *ent.SSHCommand) bo.SSHCommand {
	return bo.SSHCommand{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name, Command: row.Command,
		Workdir: row.Workdir, Env: row.Env, ReviewStatus: string(row.ReviewStatus), Remark: row.Remark,
	}
}

func toSSHAudit(row *ent.SSHCommandAudit) bo.SSHCommandAudit {
	return bo.SSHCommandAudit{
		UID: row.UID, NamespaceUID: row.NamespaceUID, CommandUID: row.CommandUID, AgentUID: row.AgentUID,
		OperatorUID: row.OperatorUID, Kind: string(row.Kind), Status: string(row.Status), Command: row.Command,
		ExitCode: row.ExitCode, Output: row.Output, Truncated: row.Truncated, Detail: row.Detail, CreatedAt: row.CreatedAt,
	}
}
