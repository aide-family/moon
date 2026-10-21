package service

import (
	"context"
	"time"

	megrezv1 "github.com/aide-family/Alioth/api/megrez/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/alertevent"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/worker"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

type EvalService struct {
	megrezv1.UnimplementedEvalServiceServer
	events  *alertevent.Usecase
	workers *worker.Usecase
}

func NewEvalService(uc *alertevent.Usecase, workers *worker.Usecase) *EvalService {
	return &EvalService{events: uc, workers: workers}
}

func (s *EvalService) ClaimRules(ctx context.Context, _ *megrezv1.ClaimRulesRequest) (*megrezv1.ClaimRulesResponse, error) {
	items, err := s.events.Claim(ctx)
	if err != nil {
		return nil, err
	}
	rules := make([]*megrezv1.RuleSnapshot, 0, len(items))
	for _, item := range items {
		rules = append(rules, snapshotItem(item))
	}
	return &megrezv1.ClaimRulesResponse{Rules: rules}, nil
}

func (s *EvalService) ReportEval(ctx context.Context, req *megrezv1.ReportEvalRequest) (*megrezv1.ReportEvalResponse, error) {
	slot, err := parseSlot(req.GetEvalSlot())
	if err != nil {
		return nil, err
	}
	hits := make([]bo.EvalHit, 0, len(req.GetHits()))
	for _, hit := range req.GetHits() {
		hits = append(hits, bo.EvalHit{
			Fingerprint: hit.GetFingerprint(), Labels: hit.GetLabels(), Value: hit.GetValue(),
			LevelUID: hit.GetLevelUid(), Summary: hit.GetSummary(), Description: hit.GetDescription(),
		})
	}
	if err := s.events.Report(ctx, bo.EvalReport{
		RuleUID: req.GetRuleUid(), DatasourceUID: req.GetDatasourceUid(),
		EvalSlot: slot.UTC(), Skipped: req.GetSkipped(), Hits: hits,
	}); err != nil {
		return nil, err
	}
	return &megrezv1.ReportEvalResponse{}, nil
}

func (s *EvalService) ReportWorkerStatus(ctx context.Context, req *megrezv1.ReportWorkerStatusRequest) (*megrezv1.ReportWorkerStatusResponse, error) {
	var startedAt *time.Time
	if req.GetBusy() {
		parsed, err := parseSlot(req.GetTaskStartedAt())
		if err != nil {
			return nil, err
		}
		startedAt = &parsed
	}
	if err := s.workers.Report(ctx, "MEGREZ", bo.WorkerStatus{
		InstanceUID: req.GetInstanceUid(), Hostname: req.GetHostname(), Busy: req.GetBusy(),
		TaskUID: req.GetTaskUid(), TaskName: req.GetTaskName(), TaskStartedAt: startedAt,
	}); err != nil {
		return nil, err
	}
	return &megrezv1.ReportWorkerStatusResponse{}, nil
}

func snapshotItem(item bo.RuleSnapshot) *megrezv1.RuleSnapshot {
	levels := make([]*megrezv1.LevelCondition, 0, len(item.Rule.Levels))
	for _, level := range item.Rule.Levels {
		levels = append(levels, &megrezv1.LevelCondition{
			Condition: level.Condition, Sample: level.Sample, Threshold: level.Threshold,
			ForSeconds: int32(level.ForSeconds), LevelUid: level.LevelUID,
		})
	}
	slot := ""
	if item.Rule.ScheduledSlot != nil {
		slot = item.Rule.ScheduledSlot.UTC().Format(time.RFC3339Nano)
	}
	return &megrezv1.RuleSnapshot{
		Uid: item.Rule.UID, NamespaceUid: item.Rule.NamespaceUID, Type: item.Rule.Type, Driver: item.Rule.Driver,
		Expr: item.Rule.Expr, Labels: item.Rule.Labels, Summary: item.Rule.Summary, Description: item.Rule.Description,
		EvalSlot: slot, DatasourceEndpoint: item.Datasource.Endpoint, DatasourceUsername: item.Datasource.Username,
		DatasourcePassword: item.Datasource.Password, DatasourceSecret: item.Datasource.Secret,
		Levels: levels, FiringFingerprints: item.FiringFingerprints, DatasourceUid: item.Rule.DatasourceUID, Name: item.Rule.Name,
	}
}
