package do

import (
	"time"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/util/timer"
)

// TimeEngine 时间引擎接口
type TimeEngine interface {
	TeamBase
	GetName() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetRules() []TimeEngineRule
	Allow(time.Time) (bool, error)
}

// TimeEngineRule 时间引擎规则接口
type TimeEngineRule interface {
	TeamBase
	GetName() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetTimeEngines() []TimeEngine
	GetType() vobj.TimeEngineRuleType
	GetRules() []int
	ToTimerMatcher() (timer.Matcher, error)
}
