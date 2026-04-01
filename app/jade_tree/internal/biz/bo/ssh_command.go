package bo

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
)

// SSHCommandFields holds editable template fields shared by create, update, and audit payloads.
type SSHCommandFields struct {
	Name        string
	Description string
	Content     string
	WorkDir     string
	Env         map[string]string
}

// SSHCommandCreateRepoBo inserts an approved command row.
type SSHCommandCreateRepoBo struct {
	Creator snowflake.ID
	Fields  SSHCommandFields
}

// SSHCommandUpdateRepoBo updates an approved command row by UID.
type SSHCommandUpdateRepoBo struct {
	UID    snowflake.ID
	Fields SSHCommandFields
}

// SSHCommandCountByNameBo counts commands with a given name, optionally excluding one UID.
type SSHCommandCountByNameBo struct {
	Name       string
	ExcludeUID snowflake.ID
}

// CommandAuditCreateRepoBo inserts a pending audit proposal.
type CommandAuditCreateRepoBo struct {
	Creator         snowflake.ID
	TargetCommandID snowflake.ID
	Kind            enum.SSHCommandAuditKind
	Fields          SSHCommandFields
}

// CommandAuditRejectBo rejects a pending audit.
type CommandAuditRejectBo struct {
	AuditUID snowflake.ID
	Reviewer snowflake.ID
	Reason   string
}

// SubmitSSHCommandUpdateInput submits a change proposal for an existing command.
type SubmitSSHCommandUpdateInput struct {
	CommandUID snowflake.ID
	Fields     SSHCommandFields
}

// ExecuteStoredSSHCommandBo runs a stored template on a remote host.
type ExecuteStoredSSHCommandBo struct {
	CommandUID     snowflake.ID
	Host           string
	Port           int
	Username       string
	Password       string
	PrivateKey     string
	TimeoutSeconds int32
}

// RejectSSHCommandAuditInput carries audit UID and reject reason (reviewer comes from context).
type RejectSSHCommandAuditInput struct {
	AuditUID snowflake.ID
	Reason   string
}

// SSHCommandItemBo is an approved SSH command definition.
type SSHCommandItemBo struct {
	UID         int64
	Name        string
	Description string
	Content     string
	WorkDir     string
	Env         map[string]string
	Disabled    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SSHCommandAuditItemBo is a create/update proposal and its review outcome.
type SSHCommandAuditItemBo struct {
	UID              int64
	TargetCommandUID int64
	Kind             enum.SSHCommandAuditKind
	Status           enum.SSHCommandAuditStatus
	Name             string
	Description      string
	Content          string
	WorkDir          string
	Env              map[string]string
	RejectReason     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ReviewedAt       *time.Time
	ReviewerUID      int64
}

// ListSSHCommandsBo lists approved commands.
type ListSSHCommandsBo struct {
	*PageRequestBo
	Keyword string
}

// ListSSHCommandAuditsBo lists audit records.
type ListSSHCommandAuditsBo struct {
	*PageRequestBo
	StatusFilter enum.SSHCommandAuditStatus
}

// NewListSSHCommandsBo normalizes pagination defaults.
func NewListSSHCommandsBo(page, pageSize int32, keyword string) *ListSSHCommandsBo {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return &ListSSHCommandsBo{
		PageRequestBo: NewPageRequestBo(page, pageSize),
		Keyword:       keyword,
	}
}

// NewListSSHCommandAuditsBo normalizes pagination defaults.
func NewListSSHCommandAuditsBo(page, pageSize int32, statusFilter enum.SSHCommandAuditStatus) *ListSSHCommandAuditsBo {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return &ListSSHCommandAuditsBo{
		PageRequestBo: NewPageRequestBo(page, pageSize),
		StatusFilter:  statusFilter,
	}
}

// ToAPIV1SSHCommandItem converts BO to API response model.
func ToAPIV1SSHCommandItem(x *SSHCommandItemBo) *apiv1.SSHCommandItem {
	if x == nil {
		return nil
	}
	env := x.Env
	if env == nil {
		env = map[string]string{}
	}
	return &apiv1.SSHCommandItem{
		Uid:         x.UID,
		Name:        x.Name,
		Description: x.Description,
		Content:     x.Content,
		WorkDir:     x.WorkDir,
		Env:         env,
		Disabled:    x.Disabled,
		CreatedAt:   timex.FormatTime(&x.CreatedAt),
		UpdatedAt:   timex.FormatTime(&x.UpdatedAt),
	}
}

// ToAPIV1SSHCommandAuditItem converts audit BO to API response model.
func ToAPIV1SSHCommandAuditItem(x *SSHCommandAuditItemBo) *apiv1.SSHCommandAuditItem {
	if x == nil {
		return nil
	}
	env := x.Env
	if env == nil {
		env = map[string]string{}
	}
	return &apiv1.SSHCommandAuditItem{
		Uid:              x.UID,
		TargetCommandUid: x.TargetCommandUID,
		Kind:             x.Kind,
		Status:           x.Status,
		Name:             x.Name,
		Description:      x.Description,
		Content:          x.Content,
		WorkDir:          x.WorkDir,
		Env:              env,
		RejectReason:     x.RejectReason,
		CreatedAt:        timex.FormatTime(&x.CreatedAt),
		UpdatedAt:        timex.FormatTime(&x.UpdatedAt),
		ReviewerUid:      x.ReviewerUID,
		ReviewedAt:       timex.FormatTime(x.ReviewedAt),
	}
}
