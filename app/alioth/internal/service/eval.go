package service

import (
	"context"
	"encoding/json"
	"time"

	megrezv1 "github.com/aide-family/moon/api/megrez/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/alertevent"
	"github.com/aide-family/moon/app/alioth/internal/biz/worker"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/probe"
)

type EvalService struct {
	megrezv1.UnimplementedEvalServiceServer
	coordinator *alertevent.Coordinator
	roster      *worker.Roster
}

func NewEvalService(coordinator *alertevent.Coordinator, roster *worker.Roster) *EvalService {
	return &EvalService{coordinator: coordinator, roster: roster}
}

func (s *EvalService) ClaimRules(ctx context.Context, _ *megrezv1.ClaimRulesRequest) (*megrezv1.ClaimRulesResponse, error) {
	items, err := s.coordinator.Claim(ctx)
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
	if err := s.coordinator.Report(ctx, bo.EvalReport{
		RuleUID: req.GetRuleUid(), DatasourceUID: req.GetDatasourceUid(),
		EvalSlot: slot.UTC(), Skipped: req.GetSkipped(), Failed: req.GetFailed(),
		FailureMessage: req.GetFailureMessage(), Hits: hits,
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
	roundStartedAt, err := parseOptionalSlot(req.GetRoundStartedAt())
	if err != nil {
		return nil, err
	}
	nextPollAt, err := parseOptionalSlot(req.GetNextPollAt())
	if err != nil {
		return nil, err
	}
	tasks := make([]bo.WorkerTask, 0, len(req.GetTasks()))
	for _, task := range req.GetTasks() {
		slot, err := parseSlot(task.GetEvalSlot())
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, bo.WorkerTask{
			State: task.GetState(), RuleUID: task.GetRuleUid(), RuleName: task.GetRuleName(),
			DatasourceUID: task.GetDatasourceUid(), DatasourceName: task.GetDatasourceName(), EvalSlot: slot,
		})
	}
	if err := s.roster.Report(ctx, "MEGREZ", bo.WorkerStatus{
		InstanceUID: req.GetInstanceUid(), Hostname: req.GetHostname(), Busy: req.GetBusy(),
		TaskUID: req.GetTaskUid(), TaskName: req.GetTaskName(), TaskStartedAt: startedAt,
		RoundStartedAt: roundStartedAt, NextPollAt: nextPollAt,
		PollIntervalSeconds: req.GetPollIntervalSeconds(), Tasks: tasks,
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
	variables := make([]*megrezv1.RuleVariable, 0, len(item.Rule.Variables))
	for _, variable := range item.Rule.Variables {
		variables = append(variables, &megrezv1.RuleVariable{
			Name: variable.Name, Expr: variable.Expr, Description: variable.Description,
		})
	}
	return &megrezv1.RuleSnapshot{
		Uid: item.Rule.UID, NamespaceUid: item.Rule.NamespaceUID, Driver: item.Rule.Driver,
		Expr: item.Rule.Expr, Labels: item.Rule.Labels, Summary: item.Rule.Summary, Description: item.Rule.Description,
		EvalSlot: slot, DatasourceEndpoint: item.Datasource.Endpoint, DatasourceUsername: item.Datasource.Username,
		DatasourcePassword: item.Datasource.Password, DatasourceSecret: item.Datasource.Secret,
		// Elasticsearch 的评测需要知道搜哪些索引、按哪个字段切时间桶。
		DatasourceIndexPattern: item.Datasource.IndexPattern, DatasourceTimeField: item.Datasource.TimeField,
		DatasourceName: item.Datasource.Name,
		Levels:         levels, FiringFingerprints: item.FiringFingerprints, DatasourceUid: item.Rule.DatasourceUID, Name: item.Rule.Name,
		Variables: variables,
	}
}

// ClaimProbeTasks 返回本轮可执行的探测任务。执行器身份一期只接受集群内。
func (s *EvalService) ClaimProbeTasks(ctx context.Context, req *megrezv1.ClaimProbeTasksRequest) (*megrezv1.ClaimProbeTasksResponse, error) {
	kind := req.GetExecutor().GetKind()
	if kind == "" {
		kind = "CLUSTER"
	}
	if kind != "CLUSTER" {
		// 边缘探测还没有实现，明确返回空任务而不是报错，方便将来接入。
		return &megrezv1.ClaimProbeTasksResponse{}, nil
	}
	tasks, err := s.coordinator.ClaimProbeTasks(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*megrezv1.ProbeTask, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, probeTaskItem(task))
	}
	return &megrezv1.ClaimProbeTasksResponse{Tasks: out}, nil
}

// ReportProbeResults 接收执行器的探测回写。
func (s *EvalService) ReportProbeResults(ctx context.Context, req *megrezv1.ReportProbeResultsRequest) (*megrezv1.ReportProbeResultsResponse, error) {
	slot, err := parseSlot(req.GetSlot())
	if err != nil {
		return nil, err
	}
	results := make([]bo.ProbeReportItem, 0, len(req.GetResults()))
	for _, item := range req.GetResults() {
		results = append(results, bo.ProbeReportItem{
			TargetUID: item.GetTargetUid(), OK: item.GetOk(), Value: item.GetValue(),
			DurationMS: int(item.GetDurationMs()), StatusCode: int(item.GetStatusCode()),
			ErrorKind: item.GetErrorKind(), ErrorMessage: item.GetErrorMessage(),
			DNSMS: int(item.GetDnsMs()), TCPMS: int(item.GetTcpMs()), TLSMS: int(item.GetTlsMs()),
			TTFBMS: int(item.GetTtfbMs()), CertDays: int(item.GetCertDays()),
			ResolvedIP: item.GetResolvedIp(), Executor: item.GetExecutor(),
		})
	}
	if err := s.coordinator.ReportProbeResults(ctx, bo.ProbeReport{
		RuleUID: req.GetRuleUid(), Slot: slot.UTC(), Results: results,
	}); err != nil {
		return nil, err
	}
	return &megrezv1.ReportProbeResultsResponse{}, nil
}

// probeTaskItem 把探测任务转成 proto 快照。
func probeTaskItem(task bo.ProbeTask) *megrezv1.ProbeTask {
	levels := make([]*megrezv1.LevelCondition, 0, len(task.Rule.Levels))
	for _, level := range task.Rule.Levels {
		levels = append(levels, &megrezv1.LevelCondition{
			Condition: level.Condition, Sample: level.Sample, Threshold: level.Threshold,
			ForSeconds: int32(level.ForSeconds), LevelUid: level.LevelUID,
		})
	}
	targets := make([]*megrezv1.ProbeTargetRef, 0, len(task.Targets))
	for _, target := range task.Targets {
		targets = append(targets, &megrezv1.ProbeTargetRef{
			Uid: target.UID, Spec: target.Spec, Labels: target.Labels,
		})
	}
	request, err := probe.ParseRequest(task.Group.Request)
	if err != nil {
		request = probe.DefaultRequest()
	}
	return &megrezv1.ProbeTask{
		Rule: &megrezv1.ProbeRuleSnapshot{
			Uid: task.Rule.UID, NamespaceUid: task.Rule.NamespaceUID, Name: task.Rule.Name,
			Driver: task.Rule.Driver, ProbeMetric: task.Rule.ProbeMetric,
			ProbeIntervalSeconds: int32(task.Rule.ProbeIntervalSeconds),
			Labels:               task.Rule.Labels, Summary: task.Rule.Summary, Description: task.Rule.Description,
			Levels: levels,
		},
		Group: &megrezv1.ProbeGroupSnapshot{
			Uid: task.Group.UID, Name: task.Group.Name, Labels: task.Group.Labels,
			Request: &megrezv1.ProbeRequestSnapshot{
				Method: request.Method, ExpectStatus: request.ExpectStatus, TimeoutMs: int32(request.TimeoutMS),
				FollowRedirects: request.FollowRedirects, SkipTlsVerify: request.SkipTLSVerify,
				VerifyCert: request.VerifyCert, Headers: encodeHeaders(request.Headers),
				BodyKeyword: request.BodyKeyword,
			},
		},
		Targets: targets,
		Slot:    task.Slot.UTC().Format(time.RFC3339),
	}
}

// encodeHeaders 把请求头编码成 JSON 文本，执行器侧再解析。
func encodeHeaders(headers map[string]string) string {
	if len(headers) == 0 {
		return ""
	}
	encoded, err := json.Marshal(headers)
	if err != nil {
		return ""
	}
	return string(encoded)
}
