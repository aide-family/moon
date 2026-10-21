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
	ListDatasourcesByMatchers(ctx context.Context, namespaceUID, driver, matchersJSON string) ([]bo.Datasource, error)
	// ListDatasourceUIDsByLevel 返回使用某个数据源等级的数据源，等级改名或删除后用于收口事件。
	ListDatasourceUIDsByLevel(ctx context.Context, namespaceUID, levelUID string) ([]string, error)
	SelectDatasources(ctx context.Context, namespaceUID string, limit int, lastUID, kind string) ([]bo.SelectItem, error)
	DatasourceStatus(ctx context.Context, item bo.Datasource) (string, error)
	DatasourceMetadata(ctx context.Context, item bo.Datasource) (*bo.DatasourceMetadata, error)
	ListMetrics(ctx context.Context, item bo.Datasource) ([]bo.DatasourceMetric, error)
	MetricDetail(ctx context.Context, item bo.Datasource, name string) (string, error)
	InstantQuery(ctx context.Context, item bo.Datasource, expr string, at int64) (string, error)
	RangeQuery(ctx context.Context, item bo.Datasource, expr string, start, end, step int64) (string, error)
	QueryLogs(ctx context.Context, item bo.Datasource, query string, start, end int64, limit int, direction string) (string, error)
	ProxyQuery(ctx context.Context, item bo.Datasource, method, path, body string) (string, error)
}

// AlertDictionaryRepo 读写告警字典。条目类型是告警级别或数据源级别。
type AlertDictionaryRepo interface {
	CreateAlertDictionary(ctx context.Context, in bo.AlertDictionary) (*bo.AlertDictionary, error)
	UpdateAlertDictionary(ctx context.Context, in bo.AlertDictionary) (*bo.AlertDictionary, error)
	UpdateAlertDictionaryStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.AlertDictionary, error)
	DeleteAlertDictionary(ctx context.Context, namespaceUID, uid string) error
	GetAlertDictionary(ctx context.Context, namespaceUID, uid string) (*bo.AlertDictionary, error)
	ListAlertDictionaries(ctx context.Context, namespaceUID string, page, pageSize int, kind, name string) ([]bo.AlertDictionary, int, error)
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
	ListRuleUIDs(ctx context.Context, namespaceUID, groupUID string) ([]string, error)
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
	// CountEvents 只统计符合条件的告警数量，不读取列表数据。
	CountEvents(ctx context.Context, namespaceUID string, filter bo.AlertEventListFilter) (int, error)
	ListRuleUIDsForReceivers(ctx context.Context, namespaceUID, memberUID string, tokens []string) ([]string, error)
	GetEvent(ctx context.Context, namespaceUID, uid string) (*bo.AlertEvent, error)
	Intervene(ctx context.Context, namespaceUID string, uids []string, userUID string) error
	// Suppress 抑制一条实时告警。until 为空表示永久抑制，不设到期时间。
	Suppress(ctx context.Context, namespaceUID, uid string, until *time.Time) error
	RecoverManual(ctx context.Context, namespaceUID string, uids []string) error
	// RecoverSource 把仍在触发的事件记为恢复。规则、数据源、级别都为空时不处理。
	RecoverSource(ctx context.Context, namespaceUID string, ruleUIDs, datasourceUIDs, levelUIDs []string) error
	Stats(ctx context.Context, namespaceUID string) (bo.AlertStats, error)
	// BackfillHistory 把主库里遗留的已恢复事件迁移到历史分片，兼容升级前的数据。
	BackfillHistory(ctx context.Context, limit int) (int, error)
}
