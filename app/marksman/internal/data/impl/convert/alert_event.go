package convert

import (
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func ToAlertEventItemBo(m *do.AlertEvent) *bo.AlertEventItemBo {
	if m == nil {
		return nil
	}
	labels := make(map[string]string)
	if m.Labels != nil {
		labels = m.Labels.Map()
	}
	return &bo.AlertEventItemBo{
		UID:                 m.ID,
		NamespaceUID:        m.NamespaceUID,
		StrategyGroupUID:    m.StrategyGroupUID,
		StrategyGroupName:   m.StrategyGroupName,
		StrategyUID:         m.StrategyUID,
		StrategyName:        m.StrategyName,
		LevelUID:            m.LevelUID,
		LevelName:           m.LevelName,
		BgColor:             m.BgColor,
		DatasourceUID:       m.DatasourceUID,
		DatasourceName:      m.DatasourceName,
		DatasourceLevelName: m.DatasourceLevelName,
		Summary:             m.Summary,
		Description:         m.Description,
		Expr:                m.Expr,
		FiredAt:             m.FiredAt,
		Value:               m.Value,
		Labels:              labels,
		EvaluatorType:       m.EvaluatorType,
		EvaluatorSnapshotID: m.EvaluatorSnapshotID,
		Status:              m.Status,
		IntervenedAt:        m.IntervenedAt,
		IntervenedBy:        m.IntervenedBy,
		IntervenedByName:    m.IntervenedByName,
		SuppressedUntilAt:   m.SuppressedUntilAt,
		SuppressedBy:        m.SuppressedBy,
		SuppressedByName:    m.SuppressedByName,
		SuppressedReason:    m.SuppressedReason,
		RecoveredAt:         m.RecoveredAt,
		RecoveredBy:         m.RecoveredBy,
		RecoveredByName:     m.RecoveredByName,
		RecoveredReason:     m.RecoveredReason,
	}
}

// ToAlertEventDo builds do.AlertEvent from bo; snapshotID is from find-or-insert of evaluator_snapshots.
func ToAlertEventDo(ev *bo.AlertEventBo, evaluatorSnapshotID snowflake.ID) *do.AlertEvent {
	if ev == nil {
		return nil
	}
	levelUID := snowflake.ID(0)
	levelUID = ev.LevelUID
	m := &do.AlertEvent{
		NamespaceUID:        ev.NamespaceUID,
		StrategyGroupUID:    ev.StrategyGroupUID,
		StrategyGroupName:   ev.StrategyGroupName,
		StrategyUID:         ev.StrategyUID,
		StrategyName:        ev.StrategyName,
		LevelUID:            levelUID,
		LevelName:           ev.LevelName,
		DatasourceUID:       ev.DatasourceUID,
		DatasourceName:      ev.DatasourceName,
		Summary:             ev.Summary,
		Description:         ev.Description,
		Expr:                ev.Expr,
		FiredAt:             ev.FiredAt,
		Value:               ev.Value,
		Labels:              safety.NewMap(ev.Labels),
		EvaluatorType:       ev.EvaluatorType,
		EvaluatorSnapshotID: evaluatorSnapshotID,
		Fingerprint:         ev.Fingerprint,
		Status:              enum.AlertEventStatus_ALERT_EVENT_STATUS_FIRING,
	}
	return m
}
