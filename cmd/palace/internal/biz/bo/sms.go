package bo

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/team"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/merr"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

type TeamSMSConfig interface {
	GetID() uint32
	GetName() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetSMSConfig() *do.SMS
	GetProviderType() vobj.SMSProviderType
}

type SaveSMSConfigRequest struct {
	smsConfig do.TeamSMSConfig
	Config    *do.SMS
	ID        uint32
	Name      string
	Remark    string
	Status    vobj.GlobalStatus
	Provider  vobj.SMSProviderType
}

func (s *SaveSMSConfigRequest) Validate() error {
	if s.ID <= 0 && validate.IsNil(s.Config) {
		return merr.ErrorParamsError("sms config is nil")
	}

	return nil
}

func (s *SaveSMSConfigRequest) GetID() uint32 {
	if s == nil {
		return 0
	}
	if validate.IsNil(s.smsConfig) {
		return s.ID
	}
	return s.smsConfig.GetID()
}

func (s *SaveSMSConfigRequest) GetName() string {
	if s == nil {
		return ""
	}
	return s.Name
}

func (s *SaveSMSConfigRequest) GetRemark() string {
	if s == nil {
		return ""
	}
	return s.Remark
}

func (s *SaveSMSConfigRequest) GetStatus() vobj.GlobalStatus {
	if s == nil {
		return vobj.GlobalStatusUnknown
	}
	if s.Status.IsUnknown() {
		return vobj.GlobalStatusEnable
	}
	return s.Status
}

func (s *SaveSMSConfigRequest) GetSMSConfig() *do.SMS {
	if s == nil {
		return nil
	}
	if s.Config == nil && s.smsConfig != nil {
		return s.smsConfig.GetSMSConfig()
	}
	return s.Config
}

func (s *SaveSMSConfigRequest) GetProviderType() vobj.SMSProviderType {
	if s == nil {
		return vobj.SMSProviderTypeUnknown
	}
	return s.Provider
}

func (s *SaveSMSConfigRequest) WithSMSConfig(smsConfig do.TeamSMSConfig) TeamSMSConfig {
	s.smsConfig = smsConfig
	return s
}

type ListSMSConfigRequest struct {
	*PaginationRequest
	Keyword  string               `json:"keyword"`
	Status   vobj.GlobalStatus    `json:"status"`
	Provider vobj.SMSProviderType `json:"provider"`
}

func (r *ListSMSConfigRequest) ToListSMSConfigListReply(configs []*team.SmsConfig) *ListSMSConfigListReply {
	return &ListSMSConfigListReply{
		PaginationReply: r.ToReply(),
		Items:           slices.Map(configs, func(config *team.SmsConfig) do.TeamSMSConfig { return config }),
	}
}

type ListSMSConfigListReply = ListReply[do.TeamSMSConfig]
