package alertrule

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

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
	item, err := uc.rules.CreateAlertRuleGroup(ctx, bo.AlertRuleGroup{NamespaceUID: ns, Name: strings.TrimSpace(name)})
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateGroup(ctx context.Context, uid, name string) (*bo.AlertRuleGroup, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.rules.UpdateAlertRuleGroup(ctx, bo.AlertRuleGroup{NamespaceUID: ns, UID: uid, Name: strings.TrimSpace(name)})
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
	item, err := uc.rules.UpdateAlertRuleStatus(ctx, ns, uid, status)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if status == "DISABLED" {
		if err := uc.events.RecoverSource(ctx, ns, []string{uid}, nil, nil); err != nil {
			return nil, uc.access.Wrap(err)
		}
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
	if !datasource.DriverMatches(in.Type, in.Driver) {
		return merr.Invalid("invalid request")
	}
	if strings.TrimSpace(in.Labels) != "" && !json.Valid([]byte(in.Labels)) {
		return merr.Invalid("invalid request")
	}
	dsLabels, err := labelsutil.NormalizeObject(in.DatasourceLabels)
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
	recoverAll, levelUIDs := changedLevelUIDs(previous.Levels, next.Levels)
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
	before, err := uc.datasources.ListDatasourcesByMatchers(ctx, namespaceUID, previous.Type, previous.DatasourceLabels)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	after, err := uc.datasources.ListDatasourcesByMatchers(ctx, namespaceUID, next.Type, next.DatasourceLabels)
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

func changedLevelUIDs(previous, next []bo.AlertRuleLevel) (bool, []string) {
	if len(next) == 0 && len(previous) > 0 {
		return true, nil
	}
	remaining := map[string]int{}
	for _, level := range next {
		remaining[levelSignature(level)]++
	}
	kept := map[string]struct{}{}
	gone := map[string]struct{}{}
	for _, level := range previous {
		signature := levelSignature(level)
		if remaining[signature] > 0 {
			remaining[signature]--
			kept[level.LevelUID] = struct{}{}
			continue
		}
		gone[level.LevelUID] = struct{}{}
	}
	uids := make([]string, 0)
	for uid := range gone {
		if _, ok := kept[uid]; ok {
			continue
		}
		uids = append(uids, uid)
	}
	return false, uids
}

func levelSignature(level bo.AlertRuleLevel) string {
	return strings.Join([]string{level.Condition, level.Sample, level.Threshold, strconv.Itoa(level.ForSeconds), level.LevelUID}, "\x00")
}
