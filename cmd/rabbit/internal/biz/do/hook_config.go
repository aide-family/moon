package do

import (
	"encoding/json"

	"github.com/moon-monitor/moon/pkg/api/rabbit/common"
	"github.com/moon-monitor/moon/pkg/merr"
	"github.com/moon-monitor/moon/pkg/plugin/cache"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

var _ cache.Object = (*HookConfig)(nil)

func NewHookConfig(url string, opts ...HookConfigOption) (*HookConfig, error) {
	if err := validate.CheckURL(url); err != nil {
		return nil, err
	}
	h := &HookConfig{Url: url}
	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
		}
	}
	return h, nil
}

func WithHookConfigOptionName(name string) HookConfigOption {
	return func(h *HookConfig) error {
		if name == "" {
			return merr.ErrorParamsError("name is empty")
		}
		h.Name = name
		return nil
	}
}

func WithHookConfigOptionApp(app common.HookAPP) HookConfigOption {
	return func(h *HookConfig) error {
		if app < 0 {
			return merr.ErrorParamsError("app is empty")
		}
		h.App = app
		return nil
	}
}

func WithHookConfigOptionSecret(secret string) HookConfigOption {
	return func(h *HookConfig) error {
		h.Secret = secret
		return nil
	}
}

func WithHookConfigOptionToken(token string) HookConfigOption {
	return func(h *HookConfig) error {
		h.Token = token
		return nil
	}
}

func WithHookConfigOptionUsername(username string) HookConfigOption {
	return func(h *HookConfig) error {
		h.Username = username
		return nil
	}
}

func WithHookConfigOptionPassword(password string) HookConfigOption {
	return func(h *HookConfig) error {
		h.Password = password
		return nil
	}
}

func WithHookConfigOptionHeaders(headers map[string]string) HookConfigOption {
	return func(h *HookConfig) error {
		h.Headers = headers
		return nil
	}
}

func WithHookConfigOptionEnable(enable bool) HookConfigOption {
	return func(h *HookConfig) error {
		h.Enable = enable
		return nil
	}
}

type HookConfig struct {
	Name     string            `json:"name"`
	App      common.HookAPP    `json:"app"`
	Url      string            `json:"url"`
	Secret   string            `json:"secret"`
	Token    string            `json:"token"`
	Username string            `json:"username"`
	Password string            `json:"password"`
	Headers  map[string]string `json:"headers"`
	Enable   bool              `json:"enable"`
}

type HookConfigOption func(h *HookConfig) error

func (h *HookConfig) GetName() string {
	if h == nil {
		return ""
	}
	return h.Name
}

func (h *HookConfig) GetApp() common.HookAPP {
	if h == nil {
		return 0
	}
	return h.App
}

func (h *HookConfig) GetUrl() string {
	if h == nil {
		return ""
	}
	return h.Url
}

func (h *HookConfig) GetSecret() string {
	if h == nil {
		return ""
	}
	return h.Secret
}

func (h *HookConfig) GetToken() string {
	if h == nil {
		return ""
	}
	return h.Token
}

func (h *HookConfig) GetUsername() string {
	if h == nil {
		return ""
	}
	return h.Username
}

func (h *HookConfig) GetPassword() string {
	if h == nil {
		return ""
	}
	return h.Password
}

func (h *HookConfig) GetHeaders() map[string]string {
	if h == nil {
		return nil
	}
	return h.Headers
}

func (h *HookConfig) GetEnable() bool {
	if h == nil {
		return false
	}
	return h.Enable
}

func (h *HookConfig) MarshalBinary() (data []byte, err error) {
	return json.Marshal(h)
}

func (h *HookConfig) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, h)
}

func (h *HookConfig) UniqueKey() string {
	return h.Name
}
