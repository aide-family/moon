package build

import (
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
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

func ToTimeRange(timeRanges []string) ([]time.Time, error) {
	if len(timeRanges) != 2 {
		return nil, merr.ErrorParams("time range must be 2")
	}

	times := make([]time.Time, 0, 2)
	for _, timeRange := range timeRanges {
		t, err := timex.Parse(timeRange)
		if err != nil {
			return nil, merr.ErrorParams("%s", err).WithMetadata(map[string]string{
				"timeRange": err.Error(),
			})
		}
		times = append(times, t)
	}
	return times, nil
}
