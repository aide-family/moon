package vobj

import (
	"fmt"

	"github.com/moon-monitor/moon/pkg/api/houyi/common"
)

func MetricDatasourceUniqueKey(driver common.MetricDatasourceDriver, teamId, id uint32) string {
	return fmt.Sprintf("team_%d:driver_%d:%d", teamId, driver, id)
}

func MetricRuleUniqueKey(teamId uint32, strategyId uint32, levelId uint32, datasourceUniqueKey string) string {
	return fmt.Sprintf("team_%d:strategy_%d:level_%d:datasource_%s", teamId, strategyId, levelId, datasourceUniqueKey)
}
