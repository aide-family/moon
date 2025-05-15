package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

type OperateLogListRequest interface {
	GetPagination() *common.PaginationRequest
	GetOperateTypes() []common.OperateType
	GetKeyword() string
	GetUserId() uint32
}

func ToOperateLogListRequest(req OperateLogListRequest) *bo.OperateLogListRequest {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.OperateLogListRequest{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		OperateTypes:      slices.Map(req.GetOperateTypes(), func(operateType common.OperateType) vobj.OperateType { return vobj.OperateType(operateType) }),
		Keyword:           req.GetKeyword(),
		UserID:            req.GetUserId(),
	}
}

func ToOperateLogItem(log do.OperateLog) *common.OperateLogItem {
	if validate.IsNil(log) {
		return nil
	}
	return &common.OperateLogItem{
		OperateLogId: log.GetID(),
		Operator:     ToUserBaseItem(log.GetCreator()),
		Type:         common.OperateType(log.GetOperateType().GetValue()),
		Module:       common.ResourceModule(log.GetOperateModule().GetValue()),
		DataId:       log.GetOperateDataID(),
		DataName:     log.GetOperateDataName(),
		OperateTime:  timex.Format(log.GetCreatedAt()),
		Title:        log.GetTitle(),
		Before:       log.GetBefore(),
		After:        log.GetAfter(),
		Ip:           log.GetIP(),
	}
}

func ToOperateLogItems(logs []do.OperateLog) []*common.OperateLogItem {
	return slices.Map(logs, ToOperateLogItem)
}
