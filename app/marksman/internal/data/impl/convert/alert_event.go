package convert

import (
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func ToAlertEventItemBo(m *do.AlertEvent, levelName string) *bo.AlertEventItemBo {
	if m == nil {
		return nil
	}
	labels := make(map[string]string)
	if m.Labels != nil {
		labels = m.Labels.Map()
	}
	return &bo.AlertEventItemBo{
		UID:                 m.ID,
		StrategyUID:         m.StrategyUID,
		NamespaceUID:        m.NamespaceUID,
		LevelUID:            m.LevelUID,
		LevelName:           levelName,
		Summary:             m.Summary,
		Description:         m.Description,
		Expr:                m.Expr,
		FiredAt:             m.FiredAt,
		Value:               m.Value,
		Labels:              labels,
		DatasourceUID:       m.DatasourceUID,
		EvaluatorType:       m.EvaluatorType,
		EvaluatorSnapshotID: m.EvaluatorSnapshotID,
		Status:              m.Status,
		IntervenedAt:        m.IntervenedAt,
		IntervenedBy:        m.IntervenedBy,
		SuppressedUntilAt:   m.SuppressedUntilAt,
		SuppressedBy:        m.SuppressedBy,
		SuppressedReason:    m.SuppressedReason,
		RecoveredAt:         m.RecoveredAt,
		RecoveredBy:         m.RecoveredBy,
		RecoveredReason:     m.RecoveredReason,
	}
}

// ToAlertEventDo builds do.AlertEvent from bo; snapshotID is from find-or-insert of evaluator_snapshots.
func ToAlertEventDo(ev *bo.AlertEventBo, strategyGroupUID snowflake.ID, evaluatorSnapshotID snowflake.ID) *do.AlertEvent {
	if ev == nil {
		return nil
	}
	levelUID := snowflake.ID(0)
	if ev.Level != nil {
		levelUID = ev.Level.UID
	}
	m := &do.AlertEvent{
		NamespaceUID:        ev.NamespaceUID,
		StrategyUID:         ev.StrategyUID,
		StrategyGroupUID:    strategyGroupUID,
		LevelUID:            levelUID,
		Summary:             ev.Summary,
		Description:         ev.Description,
		Expr:                ev.Expr,
		FiredAt:             ev.FiredAt,
		Value:               ev.Value,
		Labels:              safety.NewMap(ev.Labels),
		DatasourceUID:       ev.DatasourceUID,
		EvaluatorType:       ev.EvaluatorType,
		EvaluatorSnapshotID: evaluatorSnapshotID,
		Fingerprint:         ev.Fingerprint,
		Status:              enum.AlertEventStatus_ALERT_EVENT_STATUS_FIRING,
	}
	return m
}
