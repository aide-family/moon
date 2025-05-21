package do

import (
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type OperateLog interface {
	Creator
	GetTeamID() uint32
	GetOperation() string
	GetMenuID() uint32
	GetMenuName() string
	GetRequest() string
	GetError() string
	GetOriginRequest() string
	GetDuration() time.Duration
	GetRequestTime() time.Time
	GetReplyTime() time.Time
	GetClientIP() string
	GetUserAgent() string
	GetUserBaseInfo() string
}

type SendMessageLog interface {
	Base
	GetTeamID() uint32
	GetMessageType() vobj.MessageType
	GetMessage() string
	GetRequestID() string
	GetStatus() vobj.SendMessageStatus
	GetRetryCount() int32
	GetError() string
}
