package do

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
)

type TeamSMSConfig interface {
	TeamBase
	GetName() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetSMSConfig() *SMS
	GetProviderType() vobj.SMSProviderType
}

type SMS struct {
	AccessKeyID     string `json:"accessKeyID"`
	AccessKeySecret string `json:"accessKeySecret"`
	SignName        string `json:"signName"`
	Endpoint        string `json:"endpoint"`
}
