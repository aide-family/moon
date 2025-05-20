package bo

import (
	"net/http"
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

type AddOperateLog struct {
	Operation     string
	Request       any
	Reply         any
	Error         error
	OriginRequest *http.Request
	Duration      time.Duration
	ClientIP      string
	UserID        uint32
	TeamID        uint32
}
