package data

import (
	"context"
	"encoding/json"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	megrezv1 "github.com/aide-family/moon/api/megrez/v1"
	"github.com/aide-family/moon/app/megrez/internal/conf"
	"github.com/aide-family/moon/app/megrez/internal/data/repository"
	"github.com/aide-family/moon/app/megrez/pkg/bo"
	"github.com/aide-family/moon/pkg/grpctls"
	"github.com/aide-family/moon/pkg/merr"
)

type gateway struct {
	client megrezv1.EvalServiceClient
	key    string
}

func NewGateway(c *conf.Alioth) (repository.EvalGateway, func(), error) {
	if c == nil || c.GetGrpcAddr() == "" || c.GetServiceKey() == "" {
		return nil, nil, merr.Server("alioth grpc is empty")
	}
	dialOption, err := grpctls.ClientDialOption("ALIOTH_GRPC_TLS")
	if err != nil {
		return nil, nil, err
	}
	conn, err := grpc.NewClient(c.GetGrpcAddr(), dialOption)
	if err != nil {
		return nil, nil, merr.Server("alioth grpc is empty")
	}
	cleanup := func() { _ = conn.Close() }
	return &gateway{client: megrezv1.NewEvalServiceClient(conn), key: c.GetServiceKey()}, cleanup, nil
}

func (g *gateway) Claim(ctx context.Context) ([]bo.RuleSnapshot, error) {
	reply, err := g.client.ClaimRules(g.auth(ctx), &megrezv1.ClaimRulesRequest{})
	if err != nil {
		return nil, merr.Server("claim failed")
	}
	items := make([]bo.RuleSnapshot, 0, len(reply.GetRules()))
	for _, rule := range reply.GetRules() {
		slot, err := time.Parse(time.RFC3339, rule.GetEvalSlot())
		if err != nil {
			return nil, merr.Invalid("invalid request")
		}
		levels := make([]bo.Level, 0, len(rule.GetLevels()))
		for _, level := range rule.GetLevels() {
			levels = append(levels, bo.Level{
				Condition: level.GetCondition(), Sample: level.GetSample(), Threshold: level.GetThreshold(),
				ForSeconds: int(level.GetForSeconds()), LevelUID: level.GetLevelUid(),
			})
		}
		variables := make([]bo.Variable, 0, len(rule.GetVariables()))
		for _, variable := range rule.GetVariables() {
			variables = append(variables, bo.Variable{
				Name: variable.GetName(), Expr: variable.GetExpr(), Description: variable.GetDescription(),
			})
		}
		items = append(items, bo.RuleSnapshot{
			UID: rule.GetUid(), NamespaceUID: rule.GetNamespaceUid(), Name: rule.GetName(),
			Driver: rule.GetDriver(), Expr: rule.GetExpr(), Labels: rule.GetLabels(), Summary: rule.GetSummary(),
			Description: rule.GetDescription(), EvalSlot: slot.UTC(), DatasourceUID: rule.GetDatasourceUid(),
			Endpoint: rule.GetDatasourceEndpoint(), Username: rule.GetDatasourceUsername(),
			Password: rule.GetDatasourcePassword(), Secret: rule.GetDatasourceSecret(),
			DatasourceName: rule.GetDatasourceName(),
			IndexPattern:   rule.GetDatasourceIndexPattern(), TimeField: rule.GetDatasourceTimeField(),
			Levels: levels, Firing: rule.GetFiringFingerprints(), Variables: variables,
		})
	}
	return items, nil
}

func (g *gateway) Report(ctx context.Context, report bo.EvalReport) error {
	hits := make([]*megrezv1.EvalHit, 0, len(report.Hits))
	for _, hit := range report.Hits {
		hits = append(hits, &megrezv1.EvalHit{
			Fingerprint: hit.Fingerprint, Labels: hit.Labels, Value: hit.Value,
			LevelUid: hit.LevelUID, Summary: hit.Summary, Description: hit.Description,
		})
	}
	_, err := g.client.ReportEval(g.auth(ctx), &megrezv1.ReportEvalRequest{
		RuleUid: report.RuleUID, DatasourceUid: report.DatasourceUID,
		EvalSlot: report.EvalSlot.UTC().Format(time.RFC3339Nano),
		Skipped:  report.Skipped, Hits: hits,
		Failed: report.Failed, FailureMessage: report.FailureMessage,
	})
	if err != nil {
		return merr.Server("report failed")
	}
	return nil
}

func (g *gateway) ReportStatus(ctx context.Context, status bo.WorkerStatus) error {
	startedAt := ""
	if !status.TaskStartedAt.IsZero() {
		startedAt = status.TaskStartedAt.UTC().Format(time.RFC3339Nano)
	}
	roundStartedAt := ""
	if !status.RoundStartedAt.IsZero() {
		roundStartedAt = status.RoundStartedAt.UTC().Format(time.RFC3339Nano)
	}
	nextPollAt := ""
	if !status.NextPollAt.IsZero() {
		nextPollAt = status.NextPollAt.UTC().Format(time.RFC3339Nano)
	}
	tasks := make([]*megrezv1.WorkerTaskSnapshot, 0, len(status.Tasks))
	for _, task := range status.Tasks {
		tasks = append(tasks, &megrezv1.WorkerTaskSnapshot{
			State: task.State, RuleUid: task.RuleUID, RuleName: task.RuleName,
			DatasourceUid: task.DatasourceUID, DatasourceName: task.DatasourceName,
			EvalSlot: task.EvalSlot.UTC().Format(time.RFC3339Nano),
		})
	}
	_, err := g.client.ReportWorkerStatus(g.auth(ctx), &megrezv1.ReportWorkerStatusRequest{
		InstanceUid: status.InstanceUID, Hostname: status.Hostname, Busy: status.Busy,
		TaskUid: status.TaskUID, TaskName: status.TaskName, TaskStartedAt: startedAt,
		Tasks: tasks, RoundStartedAt: roundStartedAt, NextPollAt: nextPollAt,
		PollIntervalSeconds: status.PollIntervalSeconds,
	})
	if err != nil {
		return merr.Server("report worker status failed")
	}
	return nil
}

func (g *gateway) auth(ctx context.Context) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+g.key))
}

// NewProbeGateway 复用同一个 gRPC 连接领取探测任务并回写结果。
func NewProbeGateway(c *conf.Alioth) (repository.ProbeGateway, func(), error) {
	if c == nil || c.GetGrpcAddr() == "" || c.GetServiceKey() == "" {
		return nil, nil, merr.Server("alioth grpc is empty")
	}
	dialOption, err := grpctls.ClientDialOption("ALIOTH_GRPC_TLS")
	if err != nil {
		return nil, nil, err
	}
	conn, err := grpc.NewClient(c.GetGrpcAddr(), dialOption)
	if err != nil {
		return nil, nil, merr.Server("alioth grpc is empty")
	}
	cleanup := func() { _ = conn.Close() }
	return &gateway{client: megrezv1.NewEvalServiceClient(conn), key: c.GetServiceKey()}, cleanup, nil
}

func (g *gateway) ClaimProbeTasks(ctx context.Context, kind string) ([]bo.ProbeTask, error) {
	reply, err := g.client.ClaimProbeTasks(g.auth(ctx), &megrezv1.ClaimProbeTasksRequest{
		Executor: &megrezv1.ProbeExecutor{Kind: kind},
	})
	if err != nil {
		return nil, merr.Server("claim probe tasks failed: " + err.Error())
	}
	tasks := make([]bo.ProbeTask, 0, len(reply.GetTasks()))
	for _, task := range reply.GetTasks() {
		slot, err := time.Parse(time.RFC3339, task.GetSlot())
		if err != nil {
			return nil, merr.Invalid("invalid request")
		}
		levels := make([]bo.Level, 0, len(task.GetRule().GetLevels()))
		for _, level := range task.GetRule().GetLevels() {
			levels = append(levels, bo.Level{
				Condition: level.GetCondition(), Sample: level.GetSample(), Threshold: level.GetThreshold(),
				ForSeconds: int(level.GetForSeconds()), LevelUID: level.GetLevelUid(),
			})
		}
		headers := map[string]string{}
		if raw := task.GetGroup().GetRequest().GetHeaders(); raw != "" {
			_ = json.Unmarshal([]byte(raw), &headers)
		}
		targets := make([]bo.ProbeTarget, 0, len(task.GetTargets()))
		for _, target := range task.GetTargets() {
			targets = append(targets, bo.ProbeTarget{UID: target.GetUid(), Spec: target.GetSpec(), Labels: target.GetLabels()})
		}
		tasks = append(tasks, bo.ProbeTask{
			Rule: bo.ProbeRule{
				UID: task.GetRule().GetUid(), NamespaceUID: task.GetRule().GetNamespaceUid(),
				Name: task.GetRule().GetName(), Driver: task.GetRule().GetDriver(),
				ProbeMetric:          task.GetRule().GetProbeMetric(),
				ProbeIntervalSeconds: int(task.GetRule().GetProbeIntervalSeconds()),
				Labels:               task.GetRule().GetLabels(), Summary: task.GetRule().GetSummary(),
				Description: task.GetRule().GetDescription(), Levels: levels,
			},
			Group: bo.ProbeGroup{
				UID: task.GetGroup().GetUid(), Name: task.GetGroup().GetName(), Labels: task.GetGroup().GetLabels(),
				Request: bo.ProbeRequest{
					Method: task.GetGroup().GetRequest().GetMethod(), ExpectStatus: task.GetGroup().GetRequest().GetExpectStatus(),
					TimeoutMS:       int(task.GetGroup().GetRequest().GetTimeoutMs()),
					FollowRedirects: task.GetGroup().GetRequest().GetFollowRedirects(),
					SkipTLSVerify:   task.GetGroup().GetRequest().GetSkipTlsVerify(),
					VerifyCert:      task.GetGroup().GetRequest().GetVerifyCert(),
					Headers:         headers, BodyKeyword: task.GetGroup().GetRequest().GetBodyKeyword(),
				},
			},
			Targets: targets,
			Slot:    slot,
		})
	}
	return tasks, nil
}

func (g *gateway) ReportProbeResults(ctx context.Context, report bo.ProbeReport) error {
	results := make([]*megrezv1.ProbeResult, 0, len(report.Results))
	for _, item := range report.Results {
		results = append(results, &megrezv1.ProbeResult{
			TargetUid: item.TargetUID, Ok: item.OK, Value: item.Value,
			DurationMs: int32(item.DurationMS), StatusCode: int32(item.StatusCode),
			ErrorKind: item.ErrorKind, ErrorMessage: item.ErrorMessage,
			DnsMs: int32(item.DNSMS), TcpMs: int32(item.TCPMS), TlsMs: int32(item.TLSMS),
			TtfbMs: int32(item.TTFBMS), CertDays: int32(item.CertDays),
			ResolvedIp: item.ResolvedIP, Executor: item.Executor,
		})
	}
	_, err := g.client.ReportProbeResults(g.auth(ctx), &megrezv1.ReportProbeResultsRequest{
		RuleUid: report.RuleUID, Slot: report.Slot.UTC().Format(time.RFC3339), Results: results,
	})
	if err != nil {
		return merr.Server("report probe results failed: " + err.Error())
	}
	return nil
}
