package bo

import "time"

// AlertEscalationPolicy 是一条告警升级策略：命中的 FIRING 事件在指定分钟数后
// 按步骤向通知组生成升级通知任务。
type AlertEscalationPolicy struct {
	UID          string
	NamespaceUID string
	Name         string
	Description  string
	// Labels 是标签匹配器 JSON：键到完整匹配正则。空对象表示不限制。
	Labels         string
	LevelUIDs      []string
	DatasourceUIDs []string
	RuleUIDs       []string
	Status         string
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	// Steps 按 StepOrder 升序。
	Steps []AlertEscalationStep
}

// AlertEscalationStep 是升级策略的一步。
type AlertEscalationStep struct {
	UID          string
	NamespaceUID string
	PolicyUID    string
	// StepOrder 从 0 开始，同一策略内唯一。
	StepOrder int
	// DelayMinutes 是距事件 starts_at 的累计分钟数。
	DelayMinutes          int
	NotificationGroupUIDs []string
}

// AlertEscalationPolicyFilter 是升级策略列表的查询条件。
type AlertEscalationPolicyFilter struct {
	Name   string
	Status string
}

// AlertEventEscalation 是一条事件升级记录，供事件时间线展示。
type AlertEventEscalation struct {
	UID          string
	NamespaceUID string
	EventUID     string
	PolicyUID    string
	// PolicyName 只用于展示；策略已删除时为空。
	PolicyName             string
	StepOrder              int
	DelayMinutes           int
	NotificationGroupCount int
	DueAt                  time.Time
	Status                 string
	SentAt                 *time.Time
	CreatedAt              time.Time
}
