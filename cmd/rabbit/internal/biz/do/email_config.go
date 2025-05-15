package do

import (
	"encoding/json"

	"github.com/moon-monitor/moon/pkg/plugin/cache"
)

var _ cache.Object = (*EmailConfig)(nil)

type EmailConfig struct {
	User   string `json:"user"`
	Pass   string `json:"pass"`
	Host   string `json:"host"`
	Port   uint32 `json:"port"`
	Enable bool   `json:"enable"`
	Name   string `json:"name"`
}

func (e *EmailConfig) UniqueKey() string {
	if e == nil {
		return ""
	}
	return e.Name
}

func (e *EmailConfig) GetUser() string {
	if e == nil {
		return ""
	}
	return e.User
}

func (e *EmailConfig) GetPass() string {
	if e == nil {
		return ""
	}
	return e.Pass
}

func (e *EmailConfig) GetHost() string {
	if e == nil {
		return ""
	}
	return e.Host
}

func (e *EmailConfig) GetPort() uint32 {
	if e == nil {
		return 0
	}
	return e.Port
}

func (e *EmailConfig) GetEnable() bool {
	if e == nil {
		return false
	}
	return e.Enable
}

func (e *EmailConfig) GetName() string {
	if e == nil {
		return ""
	}
	return e.Name
}

func (e *EmailConfig) MarshalBinary() (data []byte, err error) {
	return json.Marshal(e)
}

func (e *EmailConfig) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, e)
}
