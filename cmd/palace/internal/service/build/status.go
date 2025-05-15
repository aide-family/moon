package build

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
	"github.com/moon-monitor/moon/pkg/util/slices"
)

func Statuses(statuses []common.GlobalStatus) []vobj.GlobalStatus {
	return slices.Map(statuses, func(status common.GlobalStatus) vobj.GlobalStatus {
		return vobj.GlobalStatus(status)
	})
}
