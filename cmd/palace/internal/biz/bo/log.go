package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type OperateLogListRequest struct {
	*PaginationRequest
	OperateTypes []vobj.OperateType
	Keyword      string
	UserID       uint32
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
	OperateType     vobj.OperateType
	OperateMenuID   uint32
	OperateMenuName string
	OperateDataID   uint32
	OperateDataName string
	Title           string
	Before          string
	After           string
	IP              string
}
