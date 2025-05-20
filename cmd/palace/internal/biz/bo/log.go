package bo

import (
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type OperateLogListRequest struct {
	*PaginationRequest
	Operation string
	Keyword   string
	UserID    uint32
}

func (r *OperateLogListRequest) ToListReply(logs []do.OperateLog) *OperateLogListReply {
	return &OperateLogListReply{
		PaginationReply: r.ToReply(),
		Items:           logs,
	}
}

type OperateLogListReply = ListReply[do.OperateLog]

type TeamOperateLogListReply = ListReply[do.OperateLog]

type OperateLogParams struct {
	Operation     string
	MenuName      string
	MenuID        uint32
	Request       string
	Reply         string
	Error         string
	OriginRequest string
	Duration      time.Duration
	RequestTime   time.Time
	ReplyTime     time.Time
	ClientIP      string
	UserAgent     string
	UserID        uint32
	UserBaseInfo  string
	TeamID        uint32
}
