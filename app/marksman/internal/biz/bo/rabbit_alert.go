package bo

import (
	"strconv"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

const marksmanAlertSource = "marksman"

func NewRabbitReceivePrometheusWebhookRequestFromAlertEvent(uid snowflake.ID, event *AlertEventBo) *rabbitv1.ReceivePrometheusWebhookRequest {
	if event == nil {
		return nil
	}
	return &rabbitv1.ReceivePrometheusWebhookRequest{
		Version:  "marksman/v1",
		GroupKey: event.Fingerprint,
		Status:   "firing",
		Receiver: marksmanAlertSource,
		Alerts: []*rabbitv1.PrometheusAlertItem{
			{
				Status:       "firing",
				Labels:       buildRabbitAlertLabels(uid, event.NamespaceUID, event.Fingerprint, event.StrategyGroupUID, event.StrategyGroupName, event.StrategyUID, event.StrategyName, event.LevelUID, event.LevelName, event.DatasourceUID, event.DatasourceName, event.DatasourceLevelName, event.Labels),
				Annotations:  buildRabbitAlertAnnotations(event.Summary, event.Description, event.Expr, event.Value),
				StartsAt:     event.FiredAt.Format(time.RFC3339),
				GeneratorURL: "",
				Fingerprint:  event.Fingerprint,
			},
		},
		Source: marksmanAlertSource,
	}
}

func NewRabbitReceivePrometheusWebhookRequestFromAlertEventItem(item *AlertEventItemBo) *rabbitv1.ReceivePrometheusWebhookRequest {
	if item == nil {
		return nil
	}
	status := "firing"
	if item.Status == enum.AlertEventStatus_ALERT_EVENT_STATUS_RECOVERED {
		status = "resolved"
	}
	alert := &rabbitv1.PrometheusAlertItem{
		Status:      status,
		Labels:      buildRabbitAlertLabels(item.UID, item.NamespaceUID, item.Fingerprint, item.StrategyGroupUID, item.StrategyGroupName, item.StrategyUID, item.StrategyName, item.LevelUID, item.LevelName, item.DatasourceUID, item.DatasourceName, item.DatasourceLevelName, item.Labels),
		Annotations: buildRabbitAlertAnnotations(item.Summary, item.Description, item.Expr, item.Value),
		StartsAt:    item.FiredAt.Format(time.RFC3339),
		Fingerprint: BuildRecoveredFingerprint(item),
	}
	if item.RecoveredAt != nil {
		alert.EndsAt = item.RecoveredAt.Format(time.RFC3339)
	}
	return &rabbitv1.ReceivePrometheusWebhookRequest{
		Version:  "marksman/v1",
		GroupKey: alert.GetFingerprint(),
		Status:   status,
		Receiver: marksmanAlertSource,
		Alerts:   []*rabbitv1.PrometheusAlertItem{alert},
		Source:   marksmanAlertSource,
	}
}

func buildRabbitAlertLabels(
	alertEventUID snowflake.ID,
	namespaceUID snowflake.ID,
	fingerprint string,
	strategyGroupUID snowflake.ID,
	strategyGroupName string,
	strategyUID snowflake.ID,
	strategyName string,
	levelUID snowflake.ID,
	levelName string,
	datasourceUID snowflake.ID,
	datasourceName string,
	datasourceLevelName string,
	extra map[string]string,
) map[string]string {
	labels := make(map[string]string, len(extra)+12)
	for key, value := range extra {
		labels[key] = value
	}
	setLabelIfEmpty(labels, cnst.LabelNamespaceUID, strconv.FormatInt(namespaceUID.Int64(), 10))
	setLabelIfEmpty(labels, cnst.LabelStrategyGroupUID, strconv.FormatInt(strategyGroupUID.Int64(), 10))
	setLabelIfEmpty(labels, cnst.LabelStrategyGroupName, strategyGroupName)
	setLabelIfEmpty(labels, cnst.LabelStrategyUID, strconv.FormatInt(strategyUID.Int64(), 10))
	setLabelIfEmpty(labels, cnst.LabelStrategyName, strategyName)
	setLabelIfEmpty(labels, cnst.LabelLevelUID, strconv.FormatInt(levelUID.Int64(), 10))
	setLabelIfEmpty(labels, cnst.LabelLevelName, levelName)
	setLabelIfEmpty(labels, cnst.LabelDatasourceUID, strconv.FormatInt(datasourceUID.Int64(), 10))
	setLabelIfEmpty(labels, cnst.LabelDatasourceName, datasourceName)
	setLabelIfEmpty(labels, cnst.LabelDatasourceLevelName, datasourceLevelName)
	setLabelIfEmpty(labels, cnst.LabelAlertName, strategyName)
	setLabelIfEmpty(labels, cnst.LabelSeverity, levelName)
	if alertEventUID.Int64() > 0 {
		labels[cnst.LabelAlertEventUID] = strconv.FormatInt(alertEventUID.Int64(), 10)
	}
	if fingerprint != "" {
		labels[cnst.LabelFingerprint] = fingerprint
	}
	return labels
}

func setLabelIfEmpty(labels map[string]string, key, value string) {
	if labels == nil || key == "" || value == "" {
		return
	}
	if labels[key] == "" {
		labels[key] = value
	}
}

func buildRabbitAlertAnnotations(summary, description, expr string, value float64) map[string]string {
	annotations := map[string]string{
		"summary":     summary,
		"description": description,
		"expr":        expr,
		"value":       strconv.FormatFloat(value, 'f', -1, 64),
	}
	for key, value := range annotations {
		if value == "" {
			delete(annotations, key)
		}
	}
	return annotations
}

func BuildRecoveredFingerprint(item *AlertEventItemBo) string {
	if item == nil {
		return ""
	}
	if item.Fingerprint != "" {
		return item.Fingerprint
	}
	index := timex.FormatTime(&item.FiredAt)
	return BuildAlertFingerprint(index, buildRabbitAlertLabels(item.UID, item.NamespaceUID, "", item.StrategyGroupUID, item.StrategyGroupName, item.StrategyUID, item.StrategyName, item.LevelUID, item.LevelName, item.DatasourceUID, item.DatasourceName, item.DatasourceLevelName, item.Labels))
}
