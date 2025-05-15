package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
)

func Statuses(statuses []common.GlobalStatus) []vobj.GlobalStatus {
	return slices.Map(statuses, func(status common.GlobalStatus) vobj.GlobalStatus {
		return vobj.GlobalStatus(status)
	})
}
