package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
)

// TimeEngine represents a time engine use case
type TimeEngine struct {
	timeEngineRepo     repository.TimeEngine
	timeEngineRuleRepo repository.TimeEngineRule
}

// NewTimeEngineBiz creates a new time engine use case
func NewTimeEngineBiz(timeEngineRepo repository.TimeEngine, timeEngineRuleRepo repository.TimeEngineRule) *TimeEngine {
	return &TimeEngine{
		timeEngineRepo:     timeEngineRepo,
		timeEngineRuleRepo: timeEngineRuleRepo,
	}
}

// SaveTimeEngine saves a time engine
func (t *TimeEngine) SaveTimeEngine(ctx context.Context, req *bo.SaveTimeEngineRequest) error {
	rules, err := t.timeEngineRuleRepo.Find(ctx, req.RuleIds...)
	if err != nil {
		return err
	}
	req.WithRules(rules)
	if req.TimeEngineId == 0 {
		if err := req.Validate(); err != nil {
			return err
		}
		return t.timeEngineRepo.CreateTimeEngine(ctx, req)
	}
	timeEngine, err := t.timeEngineRepo.GetTimeEngine(ctx, &bo.GetTimeEngineRequest{
		TimeEngineId: req.TimeEngineId,
	})
	if err != nil {
		return err
	}
	req.WithTimeEngine(timeEngine)
	if err := req.Validate(); err != nil {
		return err
	}
	return t.timeEngineRepo.UpdateTimeEngine(ctx, req.TimeEngineId, req)
}

// UpdateTimeEngineStatus updates the status of a time engine
func (t *TimeEngine) UpdateTimeEngineStatus(ctx context.Context, req *bo.UpdateTimeEngineStatusRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	return t.timeEngineRepo.UpdateTimeEngineStatus(ctx, req)
}

// DeleteTimeEngine deletes a time engine
func (t *TimeEngine) DeleteTimeEngine(ctx context.Context, req *bo.DeleteTimeEngineRequest) error {
	return t.timeEngineRepo.DeleteTimeEngine(ctx, req)
}

// GetTimeEngine retrieves time engine details
func (t *TimeEngine) GetTimeEngine(ctx context.Context, req *bo.GetTimeEngineRequest) (do.TimeEngine, error) {
	return t.timeEngineRepo.GetTimeEngine(ctx, req)
}

// ListTimeEngine retrieves a list of time engines
func (t *TimeEngine) ListTimeEngine(ctx context.Context, req *bo.ListTimeEngineRequest) (*bo.ListTimeEngineReply, error) {
	return t.timeEngineRepo.ListTimeEngine(ctx, req)
}

// SelectTimeEngine retrieves a list of time engines
func (t *TimeEngine) SelectTimeEngine(ctx context.Context, req *bo.SelectTimeEngineRequest) (*bo.SelectTimeEngineReply, error) {
	return t.timeEngineRepo.SelectTimeEngine(ctx, req)
}

// SaveTimeEngineRule saves a time engine rule
func (t *TimeEngine) SaveTimeEngineRule(ctx context.Context, req *bo.SaveTimeEngineRuleRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	if req.TimeEngineRuleId == 0 {
		return t.timeEngineRuleRepo.CreateTimeEngineRule(ctx, req)
	}
	return t.timeEngineRuleRepo.UpdateTimeEngineRule(ctx, req.TimeEngineRuleId, req)
}

// UpdateTimeEngineRuleStatus updates the status of a time engine rule
func (t *TimeEngine) UpdateTimeEngineRuleStatus(ctx context.Context, req *bo.UpdateTimeEngineRuleStatusRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	return t.timeEngineRuleRepo.UpdateTimeEngineRuleStatus(ctx, req)
}

// DeleteTimeEngineRule deletes a time engine rule
func (t *TimeEngine) DeleteTimeEngineRule(ctx context.Context, req *bo.DeleteTimeEngineRuleRequest) error {
	return t.timeEngineRuleRepo.DeleteTimeEngineRule(ctx, req)
}

// GetTimeEngineRule retrieves time engine rule details
func (t *TimeEngine) GetTimeEngineRule(ctx context.Context, req *bo.GetTimeEngineRuleRequest) (do.TimeEngineRule, error) {
	return t.timeEngineRuleRepo.GetTimeEngineRule(ctx, req)
}

// ListTimeEngineRule retrieves a list of time engine rules
func (t *TimeEngine) ListTimeEngineRule(ctx context.Context, req *bo.ListTimeEngineRuleRequest) (*bo.ListTimeEngineRuleReply, error) {
	return t.timeEngineRuleRepo.ListTimeEngineRule(ctx, req)
}

// SelectTimeEngineRule retrieves a list of time engine rules
func (t *TimeEngine) SelectTimeEngineRule(ctx context.Context, req *bo.SelectTimeEngineRuleRequest) (*bo.SelectTimeEngineRuleReply, error) {
	return t.timeEngineRuleRepo.SelectTimeEngineRule(ctx, req)
}
