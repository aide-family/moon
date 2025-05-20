package do

import (
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/timer"
)

// TimeEngine represents the time engine interface
type TimeEngine interface {
	TeamBase
	GetName() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetRules() []TimeEngineRule
	Allow(time.Time) (bool, error)
}

// TimeEngineRule represents the time engine rule interface
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
