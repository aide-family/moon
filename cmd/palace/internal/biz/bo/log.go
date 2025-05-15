package bo

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/system"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/team"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/util/slices"
)

type OperateLogListRequest struct {
	*PaginationRequest
	OperateTypes []vobj.OperateType `json:"operateTypes"`
	Keyword      string             `json:"keyword"`
	UserID       uint32             `json:"userId"`
}

func (r *OperateLogListRequest) ToOperateLogListReply(logs []*system.OperateLog) *OperateLogListReply {
	return &OperateLogListReply{
		PaginationReply: r.ToReply(),
		Items:           slices.Map(logs, func(log *system.OperateLog) do.OperateLog { return log }),
	}
}

func (r *OperateLogListRequest) ToTeamOperateLogListReply(logs []*team.OperateLog) *TeamOperateLogListReply {
	return &TeamOperateLogListReply{
		PaginationReply: r.ToReply(),
		Items:           slices.Map(logs, func(log *team.OperateLog) do.OperateLog { return log }),
	}
}

type OperateLogListReply = ListReply[do.OperateLog]

type TeamOperateLogListReply = ListReply[do.OperateLog]

type AddOperateLog struct {
	OperateType     vobj.OperateType    `json:"operateType"`
	OperateModule   vobj.ResourceModule `json:"operateModule"`
	OperateDataID   uint32              `json:"operateDataID"`
	OperateDataName string              `json:"operateDataName"`
	Title           string              `json:"title"`
	Before          string              `json:"before"`
	After           string              `json:"after"`
	IP              string              `json:"ip"`
}
