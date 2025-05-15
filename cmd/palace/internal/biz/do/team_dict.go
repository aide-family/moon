package do

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type TeamDict interface {
	TeamBase
	GetKey() string
	GetValue() string
	GetStatus() vobj.GlobalStatus
	GetType() vobj.DictType
	GetColor() string
	GetLang() string
}
