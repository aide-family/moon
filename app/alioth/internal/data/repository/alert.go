package repository

import (
	"context"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// DatasourceRepo 读写数据源，并向对应驱动发起查询。接口不暴露 ent。
type DatasourceRepo interface {
	CreateDatasource(ctx context.Context, in bo.Datasource) (*bo.Datasource, error)
	UpdateDatasource(ctx context.Context, in bo.Datasource) (*bo.Datasource, error)
	UpdateDatasourceStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.Datasource, error)
	DeleteDatasource(ctx context.Context, namespaceUID, uid string) error
	GetDatasource(ctx context.Context, namespaceUID, uid string) (*bo.Datasource, error)
	ListDatasources(ctx context.Context, namespaceUID string, page, pageSize int, name, kind string) ([]bo.Datasource, int, error)
	ListDatasourcesByMatchers(ctx context.Context, namespaceUID, kind, matchersJSON string) ([]bo.Datasource, error)
	SelectDatasources(ctx context.Context, namespaceUID string, limit int, lastUID, kind string) ([]bo.SelectItem, error)
	DatasourceStatus(ctx context.Context, item bo.Datasource) (string, error)
	ListMetrics(ctx context.Context, item bo.Datasource) ([]string, error)
	MetricDetail(ctx context.Context, item bo.Datasource, name string) (string, error)
	InstantQuery(ctx context.Context, item bo.Datasource, expr string, at int64) (string, error)
	RangeQuery(ctx context.Context, item bo.Datasource, expr string, start, end, step int64) (string, error)
	ProxyQuery(ctx context.Context, item bo.Datasource, method, path, body string) (string, error)
}

// AlertLevelRepo 读写告警级别和数据源级别。
type AlertLevelRepo interface {
	CreateAlertLevel(ctx context.Context, in bo.AlertLevel) (*bo.AlertLevel, error)
	UpdateAlertLevel(ctx context.Context, in bo.AlertLevel) (*bo.AlertLevel, error)
	UpdateAlertLevelStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.AlertLevel, error)
	DeleteAlertLevel(ctx context.Context, namespaceUID, uid string) error
	GetAlertLevel(ctx context.Context, namespaceUID, uid string) (*bo.AlertLevel, error)
	ListAlertLevels(ctx context.Context, namespaceUID string, page, pageSize int, kind, name string) ([]bo.AlertLevel, int, error)
}

// AlertRuleRepo 读写规则组、规则和级别条件，并给出到期候选。
type AlertRuleRepo interface {
	CreateAlertRuleGroup(ctx context.Context, in bo.AlertRuleGroup) (*bo.AlertRuleGroup, error)
	UpdateAlertRuleGroup(ctx context.Context, in bo.AlertRuleGroup) (*bo.AlertRuleGroup, error)
	UpdateAlertRuleGroupStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.AlertRuleGroup, error)
	DeleteAlertRuleGroup(ctx context.Context, namespaceUID, uid string) error
	GetAlertRuleGroup(ctx context.Context, namespaceUID, uid string) (*bo.AlertRuleGroup, error)
	ListAlertRuleGroups(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.AlertRuleGroup, int, error)
	CreateAlertRule(ctx context.Context, in bo.AlertRule) (*bo.AlertRule, error)
	UpdateAlertRule(ctx context.Context, in bo.AlertRule) (*bo.AlertRule, error)
	UpdateAlertRuleStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.AlertRule, error)
	DeleteAlertRule(ctx context.Context, namespaceUID, uid string) error
	GetAlertRule(ctx context.Context, namespaceUID, uid string) (*bo.AlertRule, error)
	ListAlertRules(ctx context.Context, namespaceUID string, page, pageSize int, groupUID, name string) ([]bo.AlertRule, int, error)
	ListEvalCandidates(ctx context.Context, now time.Time) ([]bo.AlertRule, error)
	ScheduleDue(ctx context.Context, now, next time.Time) error
	ListScheduled(ctx context.Context) ([]bo.AlertRule, error)
	GetAlertRuleByUID(ctx context.Context, uid string) (*bo.AlertRule, error)
	ClearScheduledSlot(ctx context.Context, ruleUID string, slot time.Time) error
}

// AlertEventRepo 读写告警事件、当前 FIRING 占位和评测批次。
type AlertEventRepo interface {
	ApplyReport(ctx context.Context, rule bo.AlertRule, report bo.EvalReport) error
	HasBatch(ctx context.Context, ruleUID, datasourceUID string, slot time.Time) (bool, error)
	ListFiringFingerprints(ctx context.Context, namespaceUID, ruleUID, datasourceUID string) ([]string, error)
	AllBatchesRecorded(ctx context.Context, ruleUID string, datasourceUIDs []string, slot time.Time) (bool, error)
	ListEvents(ctx context.Context, namespaceUID string, filter bo.AlertEventListFilter, page, pageSize int) ([]bo.AlertEvent, int, error)
	ListRuleUIDsForReceivers(ctx context.Context, namespaceUID string, tokens []string) ([]string, error)
	GetEvent(ctx context.Context, namespaceUID, uid string) (*bo.AlertEvent, error)
	Intervene(ctx context.Context, namespaceUID string, uids []string, userUID string) error
	Suppress(ctx context.Context, namespaceUID, uid string) error
	RecoverManual(ctx context.Context, namespaceUID string, uids []string) error
	Stats(ctx context.Context, namespaceUID string) (bo.AlertStats, error)
}
