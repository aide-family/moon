package do

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type OperateLog interface {
	Creator
	GetOperateMenuID() uint32
	GetOperateMenuName() string
	GetOperateDataID() uint32
	GetOperateDataName() string
	GetTitle() string
	GetBefore() string
	GetAfter() string
	GetIP() string
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
