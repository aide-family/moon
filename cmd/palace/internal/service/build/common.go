package build

import (
	"time"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
	"github.com/moon-monitor/moon/pkg/util/timex"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func ToPaginationReply(pagination *bo.PaginationReply) *common.PaginationReply {
	if validate.IsNil(pagination) {
		return nil
	}
	return &common.PaginationReply{
		Total:    pagination.Total,
		Page:     pagination.Page,
		PageSize: pagination.Limit,
	}
}

func ToPaginationRequest(pagination *common.PaginationRequest) *bo.PaginationRequest {
	if validate.IsNil(pagination) {
		return nil
	}
	return &bo.PaginationRequest{
		Page:  pagination.GetPage(),
		Limit: pagination.GetPageSize(),
	}
}

func ToTimeRange(timeRanges []string) []time.Time {
	if len(timeRanges) != 2 {
		return nil
	}

	times := make([]time.Time, 0, 2)
	for _, timeRange := range timeRanges {
		t, err := timex.Parse(timeRange)
		if err != nil {
			return nil
		}
		times = append(times, t)
	}
	return times
}
