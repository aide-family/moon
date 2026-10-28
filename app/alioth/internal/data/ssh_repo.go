package data

import (
	"context"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/sshcommand"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/sshcommandaudit"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/user"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/pkg/merr"
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

func (r *sshCommandRepo) ListAudits(ctx context.Context, namespaceUID, commandUID, agentUID string, page, pageSize int) ([]bo.SSHCommandAudit, int, error) {
	query := r.client.SSHCommandAudit.Query().Where(sshcommandaudit.NamespaceUIDEQ(namespaceUID))
	if commandUID != "" {
		query = query.Where(sshcommandaudit.CommandUIDEQ(commandUID))
	}
	if agentUID != "" {
		query = query.Where(sshcommandaudit.AgentUIDEQ(agentUID))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(sshcommandaudit.FieldCreatedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.SSHCommandAudit, 0, len(rows))
	operatorUIDs := make([]string, 0, len(rows))
	seenOperators := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		item := toSSHAudit(row)
		items = append(items, item)
		if item.OperatorUID != "" {
			if _, ok := seenOperators[item.OperatorUID]; !ok {
				seenOperators[item.OperatorUID] = struct{}{}
				operatorUIDs = append(operatorUIDs, item.OperatorUID)
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
