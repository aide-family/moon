package bo

import (
	"encoding/json"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/validate"
)

type TeamEmailConfig interface {
	GetID() uint32
	GetName() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetEmailConfig() *do.Email
}

// SaveEmailConfigRequest represents the request to save email configuration
type SaveEmailConfigRequest struct {
	emailConfig do.TeamEmailConfig
	Config      *do.Email
	ID          uint32
	Name        string
	Remark      string
	Status      vobj.GlobalStatus
}

func (s *SaveEmailConfigRequest) Validate() error {
	if s.ID <= 0 && validate.IsNil(s.Config) {
		return merr.ErrorParams("email config is nil")
	}
	return nil
}

func (s *SaveEmailConfigRequest) GetID() uint32 {
	if s == nil {
		return 0
	}
	if validate.IsNil(s.emailConfig) {
		return s.ID
	}
	return s.emailConfig.GetID()
}

func (s *SaveEmailConfigRequest) GetName() string {
	if s == nil {
		return ""
	}
	return s.Name
}

func (s *SaveEmailConfigRequest) GetRemark() string {
	if s == nil {
		return ""
	}
	return s.Remark
}

func (s *SaveEmailConfigRequest) GetStatus() vobj.GlobalStatus {
	if s == nil {
		return vobj.GlobalStatusUnknown
	}
	if validate.IsNil(s.emailConfig) {
		return s.Status
	}
	if s.Status.IsUnknown() {
		return vobj.GlobalStatusEnable
	}
	return s.Status
}

func (s *SaveEmailConfigRequest) GetEmailConfig() *do.Email {
	if s == nil {
		return nil
	}
	return s.Config
}

func (s *SaveEmailConfigRequest) WithEmailConfig(emailConfig do.TeamEmailConfig) TeamEmailConfig {
	s.emailConfig = emailConfig
	return s
}

type ListEmailConfigRequest struct {
	*PaginationRequest
	Keyword string            `json:"keyword"`
	Status  vobj.GlobalStatus `json:"status"`
}

func (r *ListEmailConfigRequest) ToListReply(configs []do.TeamEmailConfig) *ListEmailConfigListReply {
	return &ListEmailConfigListReply{
		PaginationReply: r.ToReply(),
		Items:           configs,
	}
}

// ListEmailConfigListReply represents the response containing multiple email configurations
type ListEmailConfigListReply = ListReply[do.TeamEmailConfig]

type SendEmailParams struct {
	Email       string `json:"email"`
	Body        string `json:"body"`
	Subject     string `json:"subject"`
	ContentType string `json:"content_type"`
	RequestID   string `json:"request_id"`
	TeamID      uint32 `json:"team_id"`
}

func (s *SendEmailParams) String() string {
	bs, _ := json.Marshal(s)
	return string(bs)
}
