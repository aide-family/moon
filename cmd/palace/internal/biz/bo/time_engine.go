package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timer"
	"github.com/aide-family/moon/pkg/util/validate"
)

// SaveTimeEngineRequest represents the request parameters for saving a time engine
type SaveTimeEngineRequest struct {
	TimeEngineID uint32
	Name         string
	Remark       string
	RuleIds      []uint32

	rules      []do.TimeEngineRule
	timeEngine do.TimeEngine
}

func (r *SaveTimeEngineRequest) Validate() error {
	if validate.IsNil(r.timeEngine) {
		return merr.ErrorParams("timeEngine is required")
	}
	if len(r.rules) != len(r.RuleIds) {
		return merr.ErrorParams("rules is not exist")
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

// DeleteTimeEngineRequest represents the request parameters for deleting a time engine
type DeleteTimeEngineRequest struct {
	TimeEngineID uint32
}

// GetTimeEngineRequest represents the request parameters for getting time engine details
type GetTimeEngineRequest struct {
	TimeEngineID uint32
}

// ListTimeEngineRequest represents the request parameters for listing time engines
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

type SelectTimeEngineRequest struct {
	*PaginationRequest
	Status  vobj.GlobalStatus
	Keyword string
}

func (r *SelectTimeEngineRequest) ToSelectReply(items []do.TimeEngine) *SelectTimeEngineReply {
	return &SelectTimeEngineReply{
		PaginationReply: r.ToReply(),
		Items: slices.Map(items, func(item do.TimeEngine) SelectItem {
			return &selectItem{
				Value:    item.GetID(),
				Label:    item.GetName(),
				Disabled: item.GetDeletedAt() > 0 || !item.GetStatus().IsEnable(),
				Extra: &selectItemExtra{
					Remark: item.GetRemark(),
				},
			}
		}),
	}
}

type SelectTimeEngineReply = ListReply[SelectItem]

// SaveTimeEngineRuleRequest represents the request parameters for saving a time engine rule
type SaveTimeEngineRuleRequest struct {
	TimeEngineRuleID uint32
	Name             string
	Remark           string
	Type             vobj.TimeEngineRuleType
	Rules            []int
}

func (r *SaveTimeEngineRuleRequest) Validate() error {
	if !r.Type.Exist() || r.Type.IsUnknown() {
		return merr.ErrorParams("type is required")
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
		return merr.ErrorParams("invalid rule type: %s", ruleType)
	}
}

// DeleteTimeEngineRuleRequest represents the request parameters for deleting a time engine rule
type DeleteTimeEngineRuleRequest struct {
	TimeEngineRuleID uint32
}

// GetTimeEngineRuleRequest represents the request parameters for getting time engine rule details
type GetTimeEngineRuleRequest struct {
	TimeEngineRuleID uint32
}

// ListTimeEngineRuleRequest represents the request parameters for listing time engine rules
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

type SelectTimeEngineRuleRequest struct {
	*PaginationRequest
	Status  vobj.GlobalStatus
	Keyword string
	Types   []vobj.TimeEngineRuleType
}

func (r *SelectTimeEngineRuleRequest) ToSelectReply(items []do.TimeEngineRule) *SelectTimeEngineRuleReply {
	return &SelectTimeEngineRuleReply{
		PaginationReply: r.ToReply(),
		Items: slices.Map(items, func(item do.TimeEngineRule) SelectItem {
			return &selectItem{
				Value:    item.GetID(),
				Label:    item.GetName(),
				Disabled: item.GetDeletedAt() > 0 || !item.GetStatus().IsEnable(),
				Extra: &selectItemExtra{
					Remark: item.GetRemark(),
					Icon:   item.GetType().String(),
					Color:  item.GetType().String(),
				},
			}
		}),
	}
}

type SelectTimeEngineRuleReply = ListReply[SelectItem]

// UpdateTimeEngineStatusRequest represents the request parameters for updating time engine status
type UpdateTimeEngineStatusRequest struct {
	TimeEngineIds []uint32
	Status        vobj.GlobalStatus
}

func (r *UpdateTimeEngineStatusRequest) Validate() error {
	if len(r.TimeEngineIds) == 0 {
		return merr.ErrorParams("timeEngineIds is required")
	}
	if !r.Status.Exist() || r.Status.IsUnknown() {
		return merr.ErrorParams("status is required")
	}
	return nil
}

// UpdateTimeEngineRuleStatusRequest represents the request parameters for updating time engine rule status
type UpdateTimeEngineRuleStatusRequest struct {
	TimeEngineRuleIds []uint32
	Status            vobj.GlobalStatus
}

func (r *UpdateTimeEngineRuleStatusRequest) Validate() error {
	if len(r.TimeEngineRuleIds) == 0 {
		return merr.ErrorParams("timeEngineRuleIds is required")
	}
	if !r.Status.Exist() || r.Status.IsUnknown() {
		return merr.ErrorParams("status is required")
	}
	return nil
}
