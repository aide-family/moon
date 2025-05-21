package team

import (
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

var _ do.OperateLog = (*OperateLog)(nil)

const tableNameOperateLog = "team_operate_logs"

type OperateLog struct {
	do.TeamModel

	Operation     string        `gorm:"column:operation;type:varchar(255);not null;default:'';comment:operation" json:"operation"`
	MenuID        uint32        `gorm:"column:menu_id;type:int unsigned;not null;default:0;comment:operation menu ID" json:"menuID"`
	MenuName      string        `gorm:"column:menu_name;type:varchar(255);not null;default:'';comment:operation menu name" json:"menuName"`
	Request       string        `gorm:"column:request;type:text;not null;;comment:request" json:"request"`
	Error         string        `gorm:"column:error;type:text;not null;comment:error" json:"error"`
	OriginRequest string        `gorm:"column:origin_request;type:text;not null;comment:origin request" json:"originRequest"`
	Duration      time.Duration `gorm:"column:duration;type:bigint;not null;default:0;comment:duration" json:"duration"`
	RequestTime   time.Time     `gorm:"column:request_time;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:request time" json:"requestTime"`
	ReplyTime     time.Time     `gorm:"column:reply_time;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:reply time" json:"replyTime"`
	ClientIP      string        `gorm:"column:client_ip;type:varchar(128);not null;default:'';comment:client IP" json:"clientIP"`
	UserAgent     string        `gorm:"column:user_agent;type:varchar(255);not null;default:'';comment:user agent" json:"userAgent"`
	UserBaseInfo  string        `gorm:"column:user_base_info;type:text;not null;comment:user base info" json:"userBaseInfo"`
}

func (o *OperateLog) TableName() string {
	return tableNameOperateLog
}

func (o *OperateLog) GetOperation() string {
	return o.Operation
}

func (o *OperateLog) GetMenuID() uint32 {
	return o.MenuID
}

func (o *OperateLog) GetMenuName() string {
	return o.MenuName
}

func (o *OperateLog) GetRequest() string {
	return o.Request
}

func (o *OperateLog) GetError() string {
	return o.Error
}

func (o *OperateLog) GetOriginRequest() string {
	return o.OriginRequest
}

func (o *OperateLog) GetDuration() time.Duration {
	return o.Duration
}

func (o *OperateLog) GetRequestTime() time.Time {
	return o.RequestTime
}

func (o *OperateLog) GetReplyTime() time.Time {
	return o.ReplyTime
}

func (o *OperateLog) GetClientIP() string {
	return o.ClientIP
}

func (o *OperateLog) GetUserAgent() string {
	return o.UserAgent
}

func (o *OperateLog) GetUserBaseInfo() string {
	return o.UserBaseInfo
}

func (o *OperateLog) GetTeamID() uint32 {
	return o.TeamID
}
