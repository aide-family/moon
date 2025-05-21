package build

import (
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToOperateLogListRequest(req *common.OperateLogListRequest) *bo.OperateLogListRequest {
	if validate.IsNil(req) {
		return nil
	}

	item := &bo.OperateLogListRequest{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		Keyword:           req.GetKeyword(),
		UserID:            req.GetUserId(),
		Operation:         req.GetOperation(),
		TimeRange:         make([]time.Time, 0, 2),
	}
	if timeRange := req.GetTimeRange(); len(timeRange) == 2 {
		start, err := timex.Parse(timeRange[0])
		if err == nil {
			item.TimeRange = append(item.TimeRange, start)
		}
		end, err := timex.Parse(timeRange[1])
		if err == nil {
			item.TimeRange = append(item.TimeRange, end)
		}
	}
	return item
}

func ToOperateLogItem(log do.OperateLog) *common.OperateLogItem {
	if validate.IsNil(log) {
		return nil
	}
	return &common.OperateLogItem{
		Operation:     log.GetOperation(),
		MenuId:        log.GetMenuID(),
		MenuName:      log.GetMenuName(),
		Request:       log.GetRequest(),
		Error:         log.GetError(),
		OriginRequest: log.GetOriginRequest(),
		Duration:      int64(log.GetDuration()),
		RequestTime:   timex.Format(log.GetRequestTime()),
		ReplyTime:     timex.Format(log.GetReplyTime()),
		ClientIP:      log.GetClientIP(),
		UserAgent:     log.GetUserAgent(),
		UserBaseInfo:  log.GetUserBaseInfo(),
		CreatedAt:     timex.Format(log.GetCreatedAt()),
		UpdatedAt:     timex.Format(log.GetUpdatedAt()),
	}
}

func ToOperateLogItems(logs []do.OperateLog) []*common.OperateLogItem {
	return slices.MapFilter(logs, func(log do.OperateLog) (*common.OperateLogItem, bool) {
		item := ToOperateLogItem(log)
		return item, validate.IsNotNil(item)
	})
}
