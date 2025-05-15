package do

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/plugin/email"
)

type Email struct {
	User string `json:"user"`
	Pass string `json:"pass"`
	Host string `json:"host"`
	Port uint32 `json:"port"`
	Name string `json:"name"`
}

type TeamEmailConfig interface {
	TeamBase
	GetName() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetEmailConfig() *Email
	email.Config
}
