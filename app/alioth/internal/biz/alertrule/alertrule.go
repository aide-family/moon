package alertrule

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/biz/datasource"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/alerttmpl"
	labelsutil "github.com/aide-family/moon/pkg/labels"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewRulebook)

// maxGroupName 与 proto 校验保持一致，避免绕过 HTTP 的调用写入超长名称。
const maxGroupName = 64

// 自定义查询变量的限制，与 proto 校验和设计文档保持一致。名字的长度和字符
// 集由 variableNamePattern 覆盖。
const (
	maxRuleVariables   = 20
	maxVariableExprLen = 4096
	maxVariableDescLen = 200
)

// variableNamePattern 是模板标识符规则：字母或下划线开头，后接最多 31 个
// 字母、数字或下划线。中文和短横线不能当模板标识符，应写进说明。
var variableNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,31}$`)

// reservedVariableNames 是模板里已有的固定快捷变量。自定义变量同名会遮蔽
// 主查询标签、采样值等基础数据，因此不允许。
var reservedVariableNames = map[string]struct{}{
	"labels": {}, "value": {}, "externalURL": {}, "externalLabels": {},
}

// Rulebook 管理告警规则与规则分组。
type Rulebook struct {
	rules        repository.AlertRuleRepo
	dictionaries repository.AlertDictionaryRepo
	datasources  repository.DatasourceRepo
	events       repository.AlertEventRepo
	access       access.Checker
}

func NewRulebook(rules repository.AlertRuleRepo, dictionaries repository.AlertDictionaryRepo, datasources repository.DatasourceRepo, events repository.AlertEventRepo, checker access.Checker) *Rulebook {
	return &Rulebook{rules: rules, dictionaries: dictionaries, datasources: datasources, events: events, access: checker}
}

func (rulebook *Rulebook) CreateGroup(ctx context.Context, name string) (*bo.AlertRuleGroup, error) {
	ns, err := rulebook.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(name)
	if err := validateGroupName(trimmed); err != nil {
		return nil, err
	}
	item, err := rulebook.rules.CreateAlertRuleGroup(ctx, bo.AlertRuleGroup{NamespaceUID: ns, Name: trimmed})
	return item, rulebook.access.Wrap(err)
}

func (rulebook *Rulebook) UpdateGroup(ctx context.Context, uid, name string) (*bo.AlertRuleGroup, error) {
	ns, err := rulebook.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(name)
	if err := validateGroupName(trimmed); err != nil {
		return nil, err
	}
	item, err := rulebook.rules.UpdateAlertRuleGroup(ctx, bo.AlertRuleGroup{NamespaceUID: ns, UID: uid, Name: trimmed})
	return item, rulebook.access.Wrap(err)
}

func (rulebook *Rulebook) UpdateGroupStatus(ctx context.Context, uid, status string) (*bo.AlertRuleGroup, error) {
	ns, err := rulebook.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	item, err := rulebook.rules.UpdateAlertRuleGroupStatus(ctx, ns, uid, status)
	if err != nil {
		return nil, rulebook.access.Wrap(err)
	}
	if status == "DISABLED" {
		ruleUIDs, listErr := rulebook.rules.ListRuleUIDs(ctx, ns, uid)
		if listErr != nil {
			return nil, rulebook.access.Wrap(listErr)
		}
		if err := rulebook.events.RecoverSource(ctx, ns, ruleUIDs, nil, nil); err != nil {
			return nil, rulebook.access.Wrap(err)
		}
	}
	return item, nil
}

func (rulebook *Rulebook) DeleteGroup(ctx context.Context, uid string) error {
	ns, err := rulebook.access.Allow(ctx, perm.AlertRuleDelete)
	if err != nil {
		return err
	}
	return rulebook.access.Wrap(rulebook.rules.DeleteAlertRuleGroup(ctx, ns, uid))
}

func (rulebook *Rulebook) ListGroups(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.AlertRuleGroup, int, int, int, error) {
	ns, err := rulebook.access.Allow(ctx, perm.AlertRuleRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := rulebook.rules.ListAlertRuleGroups(ctx, ns, p, s, strings.TrimSpace(name))
	return items, total, p, s, rulebook.access.Wrap(err)
}

func validateGroupName(name string) error {
	if name == "" {
		return merr.Invalid("invalid request")
	}
	if utf8.RuneCountInString(name) > maxGroupName {
		return merr.Invalid("name is too long")
	}
	return nil
}

// SaveRule 只写规则和级别条件，不写评测批次或告警事件。
// 停用后的规则仍留在库里，但到期候选查询要求状态为启用且类型为指标。
func (rulebook *Rulebook) CreateRule(ctx context.Context, in bo.AlertRule) (*bo.AlertRule, error) {
	ns, err := rulebook.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	if err := rulebook.prepare(ctx, ns, &in); err != nil {
		return nil, err
	}
	item, err := rulebook.rules.CreateAlertRule(ctx, in)
	return item, rulebook.access.Wrap(err)
}

func (rulebook *Rulebook) UpdateRule(ctx context.Context, in bo.AlertRule) (*bo.AlertRule, error) {
	ns, err := rulebook.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	if err := rulebook.prepare(ctx, ns, &in); err != nil {
		return nil, err
	}
	previous, err := rulebook.rules.GetAlertRule(ctx, ns, in.UID)
	if err != nil {
		return nil, rulebook.access.Wrap(err)
	}
	item, err := rulebook.rules.UpdateAlertRule(ctx, in)
	if err != nil {
		return nil, rulebook.access.Wrap(err)
	}
	if err := rulebook.recoverChangedRule(ctx, ns, previous, item); err != nil {
		return nil, err
	}
	return item, nil
}

// UpdateRuleStatus 只切换启用状态，不改下一次评测时间。停用后候选查询看不到它，重新启用后才会出现。
func (rulebook *Rulebook) UpdateRuleStatus(ctx context.Context, uid, status string) (*bo.AlertRule, error) {
	ns, err := rulebook.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	// 先收口事件再停用规则：恢复走 events.RecoverSource 会归档到历史库，
	// 如果等仓储把事件改回 RECOVERED，恢复扫描就看不到 FIRING，历史分片会漏记录。
	if status == "DISABLED" {
		if err := rulebook.events.RecoverSource(ctx, ns, []string{uid}, nil, nil); err != nil {
			return nil, rulebook.access.Wrap(err)
		}
	}
	item, err := rulebook.rules.UpdateAlertRuleStatus(ctx, ns, uid, status)
	if err != nil {
		return nil, rulebook.access.Wrap(err)
	}
	return item, nil
}

func (rulebook *Rulebook) DeleteRule(ctx context.Context, uid string) error {
	ns, err := rulebook.access.Allow(ctx, perm.AlertRuleDelete)
	if err != nil {
		return err
	}
	if err := rulebook.events.RecoverSource(ctx, ns, []string{uid}, nil, nil); err != nil {
		return rulebook.access.Wrap(err)
	}
	return rulebook.access.Wrap(rulebook.rules.DeleteAlertRule(ctx, ns, uid))
}

func (rulebook *Rulebook) GetRule(ctx context.Context, uid string) (*bo.AlertRule, error) {
	ns, err := rulebook.access.Allow(ctx, perm.AlertRuleRead)
	if err != nil {
		return nil, err
	}
	item, err := rulebook.rules.GetAlertRule(ctx, ns, uid)
	return item, rulebook.access.Wrap(err)
}

func (rulebook *Rulebook) ListRules(ctx context.Context, pageNo, pageSize int32, groupUID, name string) ([]bo.AlertRule, int, int, int, error) {
	ns, err := rulebook.access.Allow(ctx, perm.AlertRuleRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := rulebook.rules.ListAlertRules(ctx, ns, p, s, groupUID, strings.TrimSpace(name))
	return items, total, p, s, rulebook.access.Wrap(err)
}

// validateRuleTemplates 在保存前用评测期同一套模板实现做语法校验。摘要和描述
// 由 megrez 在每次评测时展开，写坏的模板不会让评测失败，而是把展开错误当成
// 告警内容显示出来，所以在写入时就应当拦住。已声明的变量名要注入模板，否则
// {{ $disk_free }} 会被判成未定义变量。
func validateRuleTemplates(in *bo.AlertRule) error {
	names := make([]string, 0, len(in.Variables))
	for _, variable := range in.Variables {
		names = append(names, variable.Name)
	}
	if err := alerttmpl.ValidateWithVars(in.Summary, "summary", names); err != nil {
		return merr.Invalid("invalid summary template")
	}
	if err := alerttmpl.ValidateWithVars(in.Description, "description", names); err != nil {
		return merr.Invalid("invalid description template")
	}
	return nil
}

// validateRuleVariables 在保存前校验自定义查询变量：数量、名字格式与唯一性、
// 保留名、长度，以及变量查询语句本身的模板语法。变量语句只能用主查询标签、
// 采样值等基础字段，不能引用其它变量，否则求值顺序无法确定。
func validateRuleVariables(in *bo.AlertRule) error {
	if len(in.Variables) > maxRuleVariables {
		return merr.Invalid("too many variables")
	}
	seen := make(map[string]struct{}, len(in.Variables))
	for _, variable := range in.Variables {
		if !variableNamePattern.MatchString(variable.Name) {
			return merr.Invalid("invalid variable name")
		}
		if _, reserved := reservedVariableNames[variable.Name]; reserved {
			return merr.Invalid("reserved variable name")
		}
		if _, exists := seen[variable.Name]; exists {
			return merr.Invalid("duplicate variable name")
		}
		seen[variable.Name] = struct{}{}
		// 查询语句是模板，空白模板展开后没有可执行语句，保存时就拒绝。
		if strings.TrimSpace(variable.Expr) == "" {
			return merr.Invalid("empty variable expr")
		}
		if utf8.RuneCountInString(variable.Expr) > maxVariableExprLen {
			return merr.Invalid("variable expr is too long")
		}
		if utf8.RuneCountInString(variable.Description) > maxVariableDescLen {
			return merr.Invalid("variable description is too long")
		}
		if err := alerttmpl.Validate(variable.Expr, variable.Name); err != nil {
			return merr.Invalid("invalid variable template")
		}
	}
	return nil
}

func (rulebook *Rulebook) prepare(ctx context.Context, namespaceUID string, in *bo.AlertRule) error {
	if in.Driver == bo.ProbeDriverHTTP {
		return rulebook.prepareProbe(ctx, namespaceUID, in)
	}
	if !datasource.ValidDriver(in.Driver) {
		return merr.Invalid("invalid request")
	}
	// 查询规则不使用探测字段，保存时归一，避免驱动切换后残留旧配置。
	in.ProbeGroupLabels = "{}"
	in.ProbeIntervalSeconds = 0
	in.ProbeMetric = bo.ProbeMetricUp
	if err := validateRuleVariables(in); err != nil {
		return err
	}
	if err := validateRuleTemplates(in); err != nil {
		return err
	}
	if strings.TrimSpace(in.Labels) != "" && !json.Valid([]byte(in.Labels)) {
		return merr.Invalid("invalid request")
	}
	dsLabels, err := labelsutil.NormalizeMatchers(in.DatasourceLabels)
	if err != nil {
		return err
	}
	in.DatasourceLabels = dsLabels
	in.DatasourceUID = ""
	if _, err := rulebook.rules.GetAlertRuleGroup(ctx, namespaceUID, in.GroupUID); err != nil {
		return rulebook.access.Wrap(err)
	}
	for _, level := range in.Levels {
		item, err := rulebook.dictionaries.GetAlertDictionary(ctx, namespaceUID, level.LevelUID)
		if err != nil {
			return rulebook.access.Wrap(err)
		}
		if item.Type != "ALERT" || item.Status != "ACTIVE" {
			return merr.Invalid("invalid request")
		}
	}
	in.NamespaceUID = namespaceUID
	in.Name = strings.TrimSpace(in.Name)
	return nil
}

// prepareProbe 校验探测规则：周期、观测指标、目标组匹配器，并清空查询字段。
func (rulebook *Rulebook) prepareProbe(ctx context.Context, namespaceUID string, in *bo.AlertRule) error {
	if err := validateRuleVariables(in); err != nil {
		return err
	}
	if err := validateRuleTemplates(in); err != nil {
		return err
	}
	switch in.ProbeIntervalSeconds {
	case 15, 30, 60, 300:
	default:
		return merr.Invalid("invalid probe interval")
	}
	switch in.ProbeMetric {
	case bo.ProbeMetricUp, bo.ProbeMetricLatencyMS, bo.ProbeMetricStatusCode, bo.ProbeMetricCertDays:
	default:
		return merr.Invalid("invalid probe metric")
	}
	matchers, err := labelsutil.NormalizeMatchers(in.ProbeGroupLabels)
	if err != nil {
		return err
	}
	in.ProbeGroupLabels = matchers
	if strings.TrimSpace(in.Labels) != "" && !json.Valid([]byte(in.Labels)) {
		return merr.Invalid("invalid request")
	}
	// 探测规则没有查询表达式，也不匹配数据源。
	in.Expr = ""
	in.DatasourceLabels = "{}"
	in.DatasourceUID = ""
	in.Variables = nil
	if _, err := rulebook.rules.GetAlertRuleGroup(ctx, namespaceUID, in.GroupUID); err != nil {
		return rulebook.access.Wrap(err)
	}
	for _, level := range in.Levels {
		item, err := rulebook.dictionaries.GetAlertDictionary(ctx, namespaceUID, level.LevelUID)
		if err != nil {
			return rulebook.access.Wrap(err)
		}
		if item.Type != "ALERT" || item.Status != "ACTIVE" {
			return merr.Invalid("invalid request")
		}
	}
	in.NamespaceUID = namespaceUID
	in.Name = strings.TrimSpace(in.Name)
	return nil
}

func (rulebook *Rulebook) recoverChangedRule(ctx context.Context, namespaceUID string, previous, next *bo.AlertRule) error {
	if previous == nil || next == nil {
		return nil
	}
	// 驱动或表达式变化会改变整条规则的评测语义，旧事件不能继续代表新配置。
	recoverAll, levelUIDs := changedLevelUIDs(previous.Levels, next.Levels)
	if previous.Driver != next.Driver || previous.Expr != next.Expr {
		recoverAll = true
	}
	// 探测规则改周期、观测指标或目标组匹配范围，旧事件的含义已经变了。
	if previous.ProbeIntervalSeconds != next.ProbeIntervalSeconds ||
		previous.ProbeMetric != next.ProbeMetric ||
		previous.ProbeGroupLabels != next.ProbeGroupLabels {
		recoverAll = true
	}
	if recoverAll {
		return rulebook.access.Wrap(rulebook.events.RecoverSource(ctx, namespaceUID, []string{next.UID}, nil, nil))
	}
	if len(levelUIDs) > 0 {
		if err := rulebook.events.RecoverSource(ctx, namespaceUID, []string{next.UID}, nil, levelUIDs); err != nil {
			return rulebook.access.Wrap(err)
		}
	}
	dropped, err := rulebook.droppedDatasources(ctx, namespaceUID, previous, next)
	if err != nil {
		return err
	}
	if len(dropped) == 0 {
		return nil
	}
	return rulebook.access.Wrap(rulebook.events.RecoverSource(ctx, namespaceUID, []string{next.UID}, dropped, nil))
}

func (rulebook *Rulebook) droppedDatasources(ctx context.Context, namespaceUID string, previous, next *bo.AlertRule) ([]string, error) {
	before, err := rulebook.datasources.ListDatasourcesByMatchers(ctx, namespaceUID, previous.Driver, previous.DatasourceLabels)
	if err != nil {
		return nil, rulebook.access.Wrap(err)
	}
	after, err := rulebook.datasources.ListDatasourcesByMatchers(ctx, namespaceUID, next.Driver, next.DatasourceLabels)
	if err != nil {
		return nil, rulebook.access.Wrap(err)
	}
	kept := map[string]struct{}{}
	for _, item := range after {
		kept[item.UID] = struct{}{}
	}
	dropped := make([]string, 0)
	for _, item := range before {
		if _, ok := kept[item.UID]; ok {
			continue
		}
		dropped = append(dropped, item.UID)
	}
	return dropped, nil
}

// changedLevelUIDs 找出条件或级别发生变化的旧级别。同一个级别标识如果阈值、
// 采样方式或持续秒数变了，也要恢复旧事件，让下一轮按新条件产生新事件；
// 否则实时列表会一直显示旧条件算出的时长和采样值。
func changedLevelUIDs(previous, next []bo.AlertRuleLevel) (bool, []string) {
	if len(next) == 0 && len(previous) > 0 {
		return true, nil
	}
	kept := make(map[string]struct{}, len(next))
	for _, level := range next {
		kept[levelSignature(level)] = struct{}{}
	}
	gone := map[string]struct{}{}
	for _, level := range previous {
		if _, ok := kept[levelSignature(level)]; ok {
			continue
		}
		gone[level.LevelUID] = struct{}{}
	}
	uids := make([]string, 0, len(gone))
	for uid := range gone {
		uids = append(uids, uid)
	}
	return false, uids
}

func levelSignature(level bo.AlertRuleLevel) string {
	return strings.Join([]string{level.Condition, level.Sample, level.Threshold, strconv.Itoa(level.ForSeconds), level.LevelUID}, "\x00")
}
