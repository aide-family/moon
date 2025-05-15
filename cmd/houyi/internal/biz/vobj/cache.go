package vobj

import (
	"github.com/moon-monitor/moon/pkg/plugin/cache"
)

const (
	DatasourceCacheKey       cache.K = "houyi:config:datasource"
	MetricRuleCacheKey       cache.K = "houyi:rule:metric"
	AlertEventCacheKey       cache.K = "houyi:event:alert"
	StrategyMetricJobLockKey cache.K = "houyi:job:strategy:metric:lock"
	AlertJobLockKey          cache.K = "houyi:job:alert:lock"
)
