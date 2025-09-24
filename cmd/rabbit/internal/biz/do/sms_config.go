package do

import (
	"encoding/json"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/plugin/cache"
)

var _ cache.Object = (*SMSConfig)(nil)

type SMSConfig struct {
	AccessKeyID     string               `json:"accessKeyId"`
	AccessKeySecret string               `json:"accessKeySecret"`
	Endpoint        string               `json:"endpoint"`
	Name            string               `json:"name"`
	SignName        string               `json:"signName"`
	Type            vobj.SMSProviderType `json:"type"`
	Enable          bool                 `json:"enable"`
}

func (s *SMSConfig) GetEnable() bool {
	if s == nil {
		return false
	}
	return s.Enable
}

func (s *SMSConfig) GetType() vobj.SMSProviderType {
	if s == nil {
		return vobj.SMSProviderTypeUnknown
	}
	return s.Type
}

func (s *SMSConfig) GetAccessKeyID() string {
	if s == nil {
		return ""
	}
	return s.AccessKeyID
}

func (s *SMSConfig) GetAccessKeySecret() string {
	if s == nil {
		return ""
	}
	return s.AccessKeySecret
}

func (s *SMSConfig) GetSignName() string {
	if s == nil {
		return ""
	}
	return s.SignName
}

func (s *SMSConfig) GetEndpoint() string {
	if s == nil {
		return ""
	}
	return s.Endpoint
}

func (s *SMSConfig) GetName() string {
	if s == nil {
		return ""
	}
	return s.Name
}

func (s *SMSConfig) MarshalBinary() (data []byte, err error) {
	return json.Marshal(s)
}

func (s *SMSConfig) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *SMSConfig) UniqueKey() string {
	if s == nil {
		return ""
	}
	return s.Name
}
