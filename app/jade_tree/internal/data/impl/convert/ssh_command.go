package convert

import (
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/data/impl/do"
)

// ToPlainEnv converts a safety map to a plain map.
func ToPlainEnv(m *safety.Map[string, string]) map[string]string {
	if m == nil {
		return nil
	}
	return m.Map()
}

// ToSafetyEnv converts a plain map to a safety map.
func ToSafetyEnv(m map[string]string) *safety.Map[string, string] {
	if m == nil {
		return safety.NewMap(map[string]string{})
	}
	return safety.NewMap(m)
}

// ToSSHCommandItemBo maps a persisted command row to BO.
func ToSSHCommandItemBo(row *do.SSHCommand) *bo.SSHCommandItemBo {
	if row == nil {
		return nil
	}
	return &bo.SSHCommandItemBo{
		UID:         row.ID.Int64(),
		Name:        row.Name,
		Description: row.Description,
		Content:     row.Content,
		WorkDir:     row.WorkDir,
		Env:         ToPlainEnv(row.Env),
		Disabled:    row.Disabled,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

// ToSSHCommandAuditItemBo maps an audit row to BO.
func ToSSHCommandAuditItemBo(row *do.SSHCommandAudit) *bo.SSHCommandAuditItemBo {
	if row == nil {
		return nil
	}
	return &bo.SSHCommandAuditItemBo{
		UID:              row.ID.Int64(),
		TargetCommandUID: row.TargetCommandID.Int64(),
		Kind:             row.Kind,
		Status:           row.Status,
		Name:             row.Name,
		Description:      row.Description,
		Content:          row.Content,
		WorkDir:          row.WorkDir,
		Env:              ToPlainEnv(row.Env),
		RejectReason:     row.RejectReason,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
		ReviewedAt:       row.ReviewedAt,
		ReviewerUID:      row.Reviewer.Int64(),
	}
}

// ToSSHCommandDO builds a new command row for insert (IDs assigned in BeforeCreate).
func ToSSHCommandDO(creator snowflake.ID, fields *bo.SSHCommandFields) *do.SSHCommand {
	if fields == nil {
		fields = &bo.SSHCommandFields{}
	}
	row := &do.SSHCommand{
		Name:        fields.Name,
		Description: fields.Description,
		Content:     fields.Content,
		WorkDir:     fields.WorkDir,
		Env:         ToSafetyEnv(fields.Env),
		Disabled:    false,
	}
	row.BaseModel.Creator = creator
	return row
}

// ToSSHCommandAuditDO builds a pending audit row.
func ToSSHCommandAuditDO(in *bo.CommandAuditCreateRepoBo) *do.SSHCommandAudit {
	if in == nil {
		return nil
	}
	row := &do.SSHCommandAudit{
		TargetCommandID: in.TargetCommandID,
		Kind:            in.Kind,
		Status:          enum.SSHCommandAuditStatus_SSHCommandAuditStatus_PENDING,
		Name:            in.Fields.Name,
		Description:     in.Fields.Description,
		Content:         in.Fields.Content,
		WorkDir:         in.Fields.WorkDir,
		Env:             ToSafetyEnv(in.Fields.Env),
	}
	row.BaseModel.Creator = in.Creator
	return row
}

// ToSSHCommandFieldsFromAudit maps persisted audit payload fields to command fields BO.
func ToSSHCommandFieldsFromAudit(aud *do.SSHCommandAudit) *bo.SSHCommandFields {
	if aud == nil {
		return &bo.SSHCommandFields{}
	}
	return &bo.SSHCommandFields{
		Name:        aud.Name,
		Description: aud.Description,
		Content:     aud.Content,
		WorkDir:     aud.WorkDir,
		Env:         ToPlainEnv(aud.Env),
	}
}
