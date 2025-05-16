package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/timer"
	"github.com/aide-family/moon/pkg/util/validate"
)

// SaveTimeEngineRequest 保存时间引擎请求参数
type SaveTimeEngineRequest struct {
	TimeEngineId uint32
	Name         string
	Remark       string
	RuleIds      []uint32

	rules      []do.TimeEngineRule
	timeEngine do.TimeEngine
}

func (r *SaveTimeEngineRequest) Validate() error {
	if validate.IsNil(r.timeEngine) {
		return merr.ErrorParamsError("timeEngine is required")
	}
	if len(r.rules) != len(r.RuleIds) {
		return merr.ErrorParamsError("rules is not exist")
	}
	return nil
}

func (r *SaveTimeEngineRequest) WithRules(rules []do.TimeEngineRule) *SaveTimeEngineRequest {
	r.rules = rules
	return r
}

func (r *SaveTimeEngineRequest) WithTimeEngine(timeEngine do.TimeEngine) *SaveTimeEngineRequest {
	r.timeEngine = timeEngine
	return r
}

func (r *SaveTimeEngineRequest) GetRules() []do.TimeEngineRule {
	return r.rules
}

func (r *SaveTimeEngineRequest) GetTimeEngine() do.TimeEngine {
	return r.timeEngine
}

// DeleteTimeEngineRequest 删除时间引擎请求参数
type DeleteTimeEngineRequest struct {
	TimeEngineId uint32
}

// GetTimeEngineRequest 获取时间引擎详情请求参数
type GetTimeEngineRequest struct {
	TimeEngineId uint32
}

// ListTimeEngineRequest 获取时间引擎列表请求参数
type ListTimeEngineRequest struct {
	*PaginationRequest
	Status  vobj.GlobalStatus
	Keyword string
}

func (r *ListTimeEngineRequest) ToListReply(items []do.TimeEngine) *ListTimeEngineReply {
	return &ListTimeEngineReply{
		PaginationReply: r.ToReply(),
		Items:           items,
	}
}

type ListTimeEngineReply = ListReply[do.TimeEngine]

// SaveTimeEngineRuleRequest 保存时间引擎规则请求参数
type SaveTimeEngineRuleRequest struct {
	TimeEngineRuleId uint32
	Name             string
	Remark           string
	Type             vobj.TimeEngineRuleType
	Rules            []int
}

func (r *SaveTimeEngineRuleRequest) Validate() error {
	if !r.Type.Exist() || r.Type.IsUnknown() {
		return merr.ErrorParamsError("type is required")
	}
	if err := validateTimeEngineRule(r.Rules, r.Type); err != nil {
		return err
	}
	return nil
}

func validateTimeEngineRule(rules []int, ruleType vobj.TimeEngineRuleType) error {
	switch ruleType {
	case vobj.TimeEngineRuleTypeHourRange:
		return timer.ValidateHourRange(rules)
	case vobj.TimeEngineRuleTypeHour:
		return timer.ValidateHour(rules)
	case vobj.TimeEngineRuleTypeHourMinuteRange:
		return timer.ValidateHourMinuteRange(rules)
	case vobj.TimeEngineRuleTypeDaysOfWeek:
		return timer.ValidateDaysOfWeek(rules)
	case vobj.TimeEngineRuleTypeDayOfMonth:
		return timer.ValidateDayOfMonth(rules)
	case vobj.TimeEngineRuleTypeMonth:
		return timer.ValidateMonth(rules)
	default:
		return merr.ErrorParamsError("invalid rule type: %s", ruleType)
	}
}

// DeleteTimeEngineRuleRequest 删除时间引擎规则请求参数
type DeleteTimeEngineRuleRequest struct {
	TimeEngineRuleId uint32
}

// GetTimeEngineRuleRequest 获取时间引擎规则详情请求参数
type GetTimeEngineRuleRequest struct {
	TimeEngineRuleId uint32
}

// ListTimeEngineRuleRequest 获取时间引擎规则列表请求参数
type ListTimeEngineRuleRequest struct {
	*PaginationRequest
	Status  vobj.GlobalStatus
	Keyword string
	Types   []vobj.TimeEngineRuleType
}

func (r *ListTimeEngineRuleRequest) ToListReply(items []do.TimeEngineRule) *ListTimeEngineRuleReply {
	return &ListTimeEngineRuleReply{
		PaginationReply: r.ToReply(),
		Items:           items,
	}
}

type ListTimeEngineRuleReply = ListReply[do.TimeEngineRule]
