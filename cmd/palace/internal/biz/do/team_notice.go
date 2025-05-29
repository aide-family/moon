package do

import (
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/kv"
)

type NoticeHook interface {
	TeamBase
	GetName() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetURL() string
	GetMethod() vobj.HTTPMethod
	GetSecret() string
	GetHeaders() []*kv.KV
	GetApp() vobj.HookApp
	GetNoticeGroups() []NoticeGroup
}

type NoticeGroup interface {
	TeamBase
	GetName() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetHooks() []NoticeHook
	GetNoticeMembers() []NoticeMember
	GetEmailConfig() TeamEmailConfig
	GetSMSConfig() TeamSMSConfig
}

type NoticeMember interface {
	TeamBase
	GetNoticeGroupID() uint32
	GetUserID() uint32
	GetNoticeType() vobj.NoticeType
	GetNoticeGroup() NoticeGroup
	GetMember() TeamMember
	GetDutyCycle() []TimeEngine
}
