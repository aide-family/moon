package data

import (
	"context"
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
			IndexPattern: rule.GetDatasourceIndexPattern(), TimeField: rule.GetDatasourceTimeField(),
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
	_, err := g.client.ReportWorkerStatus(g.auth(ctx), &megrezv1.ReportWorkerStatusRequest{
		InstanceUid: status.InstanceUID, Hostname: status.Hostname, Busy: status.Busy,
		TaskUid: status.TaskUID, TaskName: status.TaskName, TaskStartedAt: startedAt,
	})
	if err != nil {
		return merr.Server("report worker status failed")
	}
	return nil
}

func (g *gateway) auth(ctx context.Context) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+g.key))
}
