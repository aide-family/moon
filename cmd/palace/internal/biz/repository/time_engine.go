package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

// TimeEngine time engine repository interface
type TimeEngine interface {
	// CreateTimeEngine create time engine
	CreateTimeEngine(ctx context.Context, req *bo.SaveTimeEngineRequest) error
	// UpdateTimeEngine update time engine
	UpdateTimeEngine(ctx context.Context, timeEngineID uint32, req *bo.SaveTimeEngineRequest) error
	// UpdateTimeEngineStatus update time engine status
	UpdateTimeEngineStatus(ctx context.Context, req *bo.UpdateTimeEngineStatusRequest) error
	// DeleteTimeEngine delete time engine
	DeleteTimeEngine(ctx context.Context, req *bo.DeleteTimeEngineRequest) error
	// GetTimeEngine get time engine details
	GetTimeEngine(ctx context.Context, req *bo.GetTimeEngineRequest) (do.TimeEngine, error)
	// ListTimeEngine get time engine list
	ListTimeEngine(ctx context.Context, req *bo.ListTimeEngineRequest) (*bo.ListTimeEngineReply, error)
	// SelectTimeEngine get time engine list
	SelectTimeEngine(ctx context.Context, req *bo.SelectTimeEngineRequest) (*bo.SelectTimeEngineReply, error)
	// Find get time engine list
	// if ids is empty, return all time engine rules
	Find(ctx context.Context, ids ...uint32) ([]do.TimeEngine, error)
}

// TimeEngineRule time engine rule repository interface
type TimeEngineRule interface {
	// CreateTimeEngineRule create time engine rule
	CreateTimeEngineRule(ctx context.Context, req *bo.SaveTimeEngineRuleRequest) error
	// UpdateTimeEngineRule update time engine rule
	UpdateTimeEngineRule(ctx context.Context, timeEngineRuleID uint32, req *bo.SaveTimeEngineRuleRequest) error
	// UpdateTimeEngineRuleStatus update time engine rule status
	UpdateTimeEngineRuleStatus(ctx context.Context, req *bo.UpdateTimeEngineRuleStatusRequest) error
	// DeleteTimeEngineRule delete time engine rule
	DeleteTimeEngineRule(ctx context.Context, req *bo.DeleteTimeEngineRuleRequest) error
	// GetTimeEngineRule get time engine rule details
	GetTimeEngineRule(ctx context.Context, req *bo.GetTimeEngineRuleRequest) (do.TimeEngineRule, error)
	// ListTimeEngineRule get time engine rule list
	ListTimeEngineRule(ctx context.Context, req *bo.ListTimeEngineRuleRequest) (*bo.ListTimeEngineRuleReply, error)
	// SelectTimeEngineRule get time engine rule list
	SelectTimeEngineRule(ctx context.Context, req *bo.SelectTimeEngineRuleRequest) (*bo.SelectTimeEngineRuleReply, error)
	// Find get time engine rule list
	// if ruleIds is empty, return all time engine rules
	Find(ctx context.Context, ruleIds ...uint32) ([]do.TimeEngineRule, error)
}
