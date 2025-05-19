package bo

import (
	"github.com/aide-family/moon/pkg/api/rabbit/common"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
)

func NewSendHookParams(configs []HookConfig, opts ...SendHookParamsOption) (SendHookParams, error) {
	configs = slices.MapFilter(configs, func(configItem HookConfig) (HookConfig, bool) {
		return configItem, configItem != nil && configItem.GetEnable()
	})
	if len(configs) == 0 {
		return nil, merr.ErrorParams("No hook configuration is available")
	}
	params := &sendHookParams{
		Configs: configs,
	}
	for _, opt := range opts {
		if err := opt(params); err != nil {
			return nil, err
		}
	}
	return params, nil
}

type HookConfig interface {
	GetName() string
	GetApp() common.HookAPP
	GetUrl() string
	GetSecret() string
	GetToken() string
	GetUsername() string
	GetPassword() string
	GetHeaders() map[string]string
	GetEnable() bool
}

type GetHookConfigParams struct {
	TeamID            string
	Name              *string
	DefaultHookConfig HookConfig
}

type SetHookConfigParams struct {
	TeamID  string
	Configs []HookConfig
}

type SendHookParams interface {
	GetBody() []*HookBody
	GetConfigs() []HookConfig
}

type sendHookParams struct {
	Body    []*HookBody
	Configs []HookConfig
}

func (s *sendHookParams) GetBody() []*HookBody {
	if s == nil {
		return nil
	}
	return s.Body
}

func (s *sendHookParams) GetConfigs() []HookConfig {
	if s == nil {
		return nil
	}
	return slices.UniqueWithFunc(s.Configs, func(configItem HookConfig) string { return configItem.GetUrl() })
}

type SendHookParamsOption func(params *sendHookParams) error

type HookBody struct {
	AppName string
	Body    []byte
}

func WithSendHookParamsOptionBody(body []*HookBody) SendHookParamsOption {
	return func(params *sendHookParams) error {
		if len(body) == 0 {
			return merr.ErrorParams("body is empty").WithMetadata(map[string]string{
				"body": "body is empty",
			})
		}
		params.Body = body
		return nil
	}
}
