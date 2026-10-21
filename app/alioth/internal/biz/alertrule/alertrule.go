package alertrule

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/datasource"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	labelsutil "github.com/aide-family/Alioth/pkg/labels"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewUsecase)

// maxGroupName 与 proto 校验保持一致，避免绕过 HTTP 的调用写入超长名称。
const maxGroupName = 64

type Usecase struct {
	rules        repository.AlertRuleRepo
	dictionaries repository.AlertDictionaryRepo
	datasources  repository.DatasourceRepo
	events       repository.AlertEventRepo
	access       access.Checker
}

func NewUsecase(rules repository.AlertRuleRepo, dictionaries repository.AlertDictionaryRepo, datasources repository.DatasourceRepo, events repository.AlertEventRepo, checker access.Checker) *Usecase {
	return &Usecase{rules: rules, dictionaries: dictionaries, datasources: datasources, events: events, access: checker}
}

func (uc *Usecase) CreateGroup(ctx context.Context, name string) (*bo.AlertRuleGroup, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(name)
	if err := validateGroupName(trimmed); err != nil {
		return nil, err
	}
	item, err := uc.rules.CreateAlertRuleGroup(ctx, bo.AlertRuleGroup{NamespaceUID: ns, Name: trimmed})
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateGroup(ctx context.Context, uid, name string) (*bo.AlertRuleGroup, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(name)
	if err := validateGroupName(trimmed); err != nil {
		return nil, err
	}
	item, err := uc.rules.UpdateAlertRuleGroup(ctx, bo.AlertRuleGroup{NamespaceUID: ns, UID: uid, Name: trimmed})
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateGroupStatus(ctx context.Context, uid, status string) (*bo.AlertRuleGroup, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.rules.UpdateAlertRuleGroupStatus(ctx, ns, uid, status)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if status == "DISABLED" {
		ruleUIDs, listErr := uc.rules.ListRuleUIDs(ctx, ns, uid)
		if listErr != nil {
			return nil, uc.access.Wrap(listErr)
		}
		if err := uc.events.RecoverSource(ctx, ns, ruleUIDs, nil, nil); err != nil {
			return nil, uc.access.Wrap(err)
		}
	}
	return item, nil
}

func (uc *Usecase) DeleteGroup(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.rules.DeleteAlertRuleGroup(ctx, ns, uid))
}

func (uc *Usecase) ListGroups(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.AlertRuleGroup, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.rules.ListAlertRuleGroups(ctx, ns, p, s, strings.TrimSpace(name))
	return items, total, p, s, uc.access.Wrap(err)
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
func (uc *Usecase) CreateRule(ctx context.Context, in bo.AlertRule) (*bo.AlertRule, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	if err := uc.prepare(ctx, ns, &in); err != nil {
		return nil, err
	}
	item, err := uc.rules.CreateAlertRule(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateRule(ctx context.Context, in bo.AlertRule) (*bo.AlertRule, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	if err := uc.prepare(ctx, ns, &in); err != nil {
		return nil, err
	}
	previous, err := uc.rules.GetAlertRule(ctx, ns, in.UID)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	item, err := uc.rules.UpdateAlertRule(ctx, in)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if err := uc.recoverChangedRule(ctx, ns, previous, item); err != nil {
		return nil, err
	}
	return item, nil
}

// UpdateRuleStatus 只切换启用状态，不改下一次评测时间。停用后候选查询看不到它，重新启用后才会出现。
func (uc *Usecase) UpdateRuleStatus(ctx context.Context, uid, status string) (*bo.AlertRule, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	// 先收口事件再停用规则：恢复走 events.RecoverSource 会归档到历史库，
	// 如果等仓储把事件改回 RECOVERED，恢复扫描就看不到 FIRING，历史分片会漏记录。
	if status == "DISABLED" {
		if err := uc.events.RecoverSource(ctx, ns, []string{uid}, nil, nil); err != nil {
			return nil, uc.access.Wrap(err)
		}
	}
	item, err := uc.rules.UpdateAlertRuleStatus(ctx, ns, uid, status)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	return item, nil
}

func (uc *Usecase) DeleteRule(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return err
	}
	if err := uc.events.RecoverSource(ctx, ns, []string{uid}, nil, nil); err != nil {
		return uc.access.Wrap(err)
	}
	return uc.access.Wrap(uc.rules.DeleteAlertRule(ctx, ns, uid))
}

func (uc *Usecase) GetRule(ctx context.Context, uid string) (*bo.AlertRule, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.rules.GetAlertRule(ctx, ns, uid)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) ListRules(ctx context.Context, pageNo, pageSize int32, groupUID, name string) ([]bo.AlertRule, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.rules.ListAlertRules(ctx, ns, p, s, groupUID, strings.TrimSpace(name))
	return items, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) prepare(ctx context.Context, namespaceUID string, in *bo.AlertRule) error {
	if !datasource.ValidDriver(in.Driver) {
		return merr.Invalid("invalid request")
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
	if _, err := uc.rules.GetAlertRuleGroup(ctx, namespaceUID, in.GroupUID); err != nil {
		return uc.access.Wrap(err)
	}
	for _, level := range in.Levels {
		item, err := uc.dictionaries.GetAlertDictionary(ctx, namespaceUID, level.LevelUID)
		if err != nil {
			return uc.access.Wrap(err)
		}
		if item.Type != "ALERT" || item.Status != "ACTIVE" {
			return merr.Invalid("invalid request")
		}
	}
	in.NamespaceUID = namespaceUID
	in.Name = strings.TrimSpace(in.Name)
	return nil
}

func (uc *Usecase) recoverChangedRule(ctx context.Context, namespaceUID string, previous, next *bo.AlertRule) error {
	if previous == nil || next == nil {
		return nil
	}
	// 驱动或表达式变化会改变整条规则的评测语义，旧事件不能继续代表新配置。
	recoverAll, levelUIDs := changedLevelUIDs(previous.Levels, next.Levels)
	if previous.Driver != next.Driver || previous.Expr != next.Expr {
		recoverAll = true
	}
	if recoverAll {
		return uc.access.Wrap(uc.events.RecoverSource(ctx, namespaceUID, []string{next.UID}, nil, nil))
	}
	if len(levelUIDs) > 0 {
		if err := uc.events.RecoverSource(ctx, namespaceUID, []string{next.UID}, nil, levelUIDs); err != nil {
			return uc.access.Wrap(err)
		}
	}
	dropped, err := uc.droppedDatasources(ctx, namespaceUID, previous, next)
	if err != nil {
		return err
	}
	if len(dropped) == 0 {
		return nil
	}
	return uc.access.Wrap(uc.events.RecoverSource(ctx, namespaceUID, []string{next.UID}, dropped, nil))
}

func (uc *Usecase) droppedDatasources(ctx context.Context, namespaceUID string, previous, next *bo.AlertRule) ([]string, error) {
	before, err := uc.datasources.ListDatasourcesByMatchers(ctx, namespaceUID, previous.Driver, previous.DatasourceLabels)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	after, err := uc.datasources.ListDatasourcesByMatchers(ctx, namespaceUID, next.Driver, next.DatasourceLabels)
	if err != nil {
		return nil, uc.access.Wrap(err)
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
