package bo

import (
	"github.com/aide-family/moon/pkg/api/rabbit/common"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewSendSMSParams(config SMSConfig, opts ...SendSMSParamsOption) (SendSMSParams, error) {
	if validate.IsNil(config) {
		return nil, merr.ErrorParams("No SMS configuration is available")
	}
	s := &sendSMSParams{
		config: config,
	}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

type SMSConfig interface {
	GetType() common.SMSConfig_Type
	GetAccessKeyId() string
	GetAccessKeySecret() string
	GetSignName() string
	GetEndpoint() string
	GetName() string
	GetEnable() bool
}

type GetSMSConfigParams struct {
	TeamID uint32
	Name   string
}

type SetSMSConfigParams struct {
	TeamID  uint32
	Configs []SMSConfig
}

type SendSMSParams interface {
	GetPhoneNumbers() []string
	GetGetTemplateParam() string
	GetTemplateCode() string
	GetConfig() SMSConfig
}

type sendSMSParams struct {
	PhoneNumbers  []string
	TemplateParam string
	TemplateCode  string

	config SMSConfig
}

func (s *sendSMSParams) GetConfig() SMSConfig {
	if s == nil {
		return nil
	}
	return s.config
}

func (s *sendSMSParams) GetPhoneNumbers() []string {
	if s == nil {
		return nil
	}
	return slices.Unique(s.PhoneNumbers)
}

func (s *sendSMSParams) GetGetTemplateParam() string {
	if s == nil {
		return ""
	}
	return s.TemplateParam
}

func (s *sendSMSParams) GetTemplateCode() string {
	if s == nil {
		return ""
	}
	return s.TemplateCode
}

type SendSMSParamsOption func(s *sendSMSParams) error

func WithSendSMSParamsOptionPhoneNumbers(phoneNumbers ...string) SendSMSParamsOption {
	return func(s *sendSMSParams) error {
		if len(phoneNumbers) == 0 {
			return merr.ErrorParams("Phone numbers cannot be empty").WithMetadata(map[string]string{
				"phoneNumbers": "Phone numbers cannot be empty",
			})
		}
		s.PhoneNumbers = phoneNumbers
		return nil
	}
}

func WithSendSMSParamsOptionTemplateParam(templateParam string) SendSMSParamsOption {
	return func(s *sendSMSParams) error {
		if templateParam == "" {
			return merr.ErrorParams("Template parameters cannot be empty").WithMetadata(map[string]string{
				"templateParam": "Template parameters cannot be empty",
			})
		}
		s.TemplateParam = templateParam
		return nil
	}
}

func WithSendSMSParamsOptionTemplateCode(templateCode string) SendSMSParamsOption {
	return func(s *sendSMSParams) error {
		if templateCode == "" {
			return merr.ErrorParams("Template code cannot be empty").WithMetadata(map[string]string{
				"templateCode": "Template code cannot be empty",
			})
		}
		s.TemplateCode = templateCode
		return nil
	}
}
