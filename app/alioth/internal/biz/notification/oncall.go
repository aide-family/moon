package notification

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/oncall"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	labelsutil "github.com/aide-family/Alioth/pkg/labels"
	"github.com/aide-family/Alioth/pkg/merr"
)

const (
	// maxOnCallFilterValues 是标签、级别和数据源条件单次允许的取值数量。
	maxOnCallFilterValues = 200
	// maxOnCallGroups 是一个 OnCall 允许绑定的通知组数量。
	maxOnCallGroups = 50
	// maxOnCallNameRunes 是名称允许的字符数。
	maxOnCallNameRunes = 64
	// maxOnCallLabelsBytes 是标签匹配器 JSON 允许的字节数。
	maxOnCallLabelsBytes = 4096
)

func (center *Center) CreateOnCall(ctx context.Context, in bo.NotificationOnCall) (*bo.NotificationOnCall, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := normalizeOnCall(&in); err != nil {
		return nil, err
	}
	if err := center.checkOnCallGroups(ctx, in.NamespaceUID, in.Groups); err != nil {
		return nil, err
	}
	item, err := center.repo.CreateOnCall(ctx, in)
	return item, center.access.Wrap(err)
}

func (center *Center) UpdateOnCall(ctx context.Context, in bo.NotificationOnCall) (*bo.NotificationOnCall, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := normalizeOnCall(&in); err != nil {
		return nil, err
	}
	if err := center.checkOnCallGroups(ctx, in.NamespaceUID, in.Groups); err != nil {
		return nil, err
	}
	item, err := center.repo.UpdateOnCall(ctx, in)
	return item, center.access.Wrap(err)
}

func (center *Center) UpdateOnCallStatus(ctx context.Context, uid, status string) (*bo.NotificationOnCall, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	item, err := center.repo.UpdateOnCallStatus(ctx, ns, uid, status)
	return item, center.access.Wrap(err)
}

func (center *Center) DeleteOnCall(ctx context.Context, uid string) error {
	ns, err := center.access.Allow(ctx, perm.NotificationDelete)
	if err != nil {
		return err
	}
	return center.access.Wrap(center.repo.DeleteOnCall(ctx, ns, uid))
}

func (center *Center) GetOnCall(ctx context.Context, uid string) (*bo.NotificationOnCall, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, err
	}
	item, err := center.repo.GetOnCall(ctx, ns, uid)
	return item, center.access.Wrap(err)
}

func (center *Center) ListOnCalls(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.NotificationOnCall, int, int, int, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := center.repo.ListOnCalls(ctx, ns, p, s, strings.TrimSpace(name))
	return items, total, p, s, center.access.Wrap(err)
}

// normalizeOnCall 校验并规整 OnCall 的条件、时区、通知组和时段。
// 只做纯校验和规范化，不访问仓储和权限，便于单测；
// 通知组是否存在由 checkOnCallGroups 依赖当前 Namespace 校验。
func normalizeOnCall(in *bo.NotificationOnCall) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return merr.Invalid("name is required")
	}
	if utf8.RuneCountInString(in.Name) > maxOnCallNameRunes {
		return merr.Invalid("name is too long")
	}
	in.LevelUIDs = uniqueStrings(in.LevelUIDs)
	in.DatasourceUIDs = uniqueStrings(in.DatasourceUIDs)
	in.DatasourceLevelUIDs = uniqueStrings(in.DatasourceLevelUIDs)
	if len(in.LevelUIDs) > maxOnCallFilterValues || len(in.DatasourceUIDs) > maxOnCallFilterValues || len(in.DatasourceLevelUIDs) > maxOnCallFilterValues {
		return merr.Invalid("too many on-call filters")
	}
	if err := validateRepeatInterval(in.RepeatIntervalSeconds); err != nil {
		return err
	}
	normalizedLabels, err := labelsutil.NormalizeMatchers(in.Labels)
	if err != nil {
		return err
	}
	if len(normalizedLabels) > maxOnCallLabelsBytes {
		return merr.Invalid("labels is too long")
	}
	in.Labels = normalizedLabels
	timezone, err := oncall.NormalizeTimezone(in.Timezone)
	if err != nil {
		return err
	}
	in.Timezone = timezone
	if len(in.Groups) == 0 {
		return merr.Invalid("notification group is required")
	}
	if len(in.Groups) > maxOnCallGroups {
		return merr.Invalid("too many on-call groups")
	}
	groups := make([]bo.NotificationOnCallGroup, 0, len(in.Groups))
	seenGroups := make(map[string]struct{}, len(in.Groups))
	for _, group := range in.Groups {
		groupUID := strings.TrimSpace(group.GroupUID)
		if groupUID == "" {
			return merr.Invalid("notification group is required")
		}
		if _, ok := seenGroups[groupUID]; ok {
			return merr.Invalid("duplicate notification group")
		}
		seenGroups[groupUID] = struct{}{}
		// 时段按 OnCall 保存，判断时再编译；这里先规范化，非法值和重复窗口直接拒绝/合并。
		windows, err := oncall.Normalize(onCallSpecs(group.Windows))
		if err != nil {
			return err
		}
		groups = append(groups, bo.NotificationOnCallGroup{
			GroupUID: groupUID,
			Windows:  onCallWindows(windows),
		})
	}
	in.Groups = groups
	return nil
}

// checkOnCallGroups 确认每个绑定的通知组都在当前 Namespace 内存在。
func (center *Center) checkOnCallGroups(ctx context.Context, namespaceUID string, groups []bo.NotificationOnCallGroup) error {
	for _, group := range groups {
		if _, err := center.repo.GetGroup(ctx, namespaceUID, group.GroupUID); err != nil {
			return center.access.Wrap(err)
		}
	}
	return nil
}

// onCallSpecs 把业务对象的时段转成 oncall 包能校验和编码的形态。
func onCallSpecs(windows []bo.NotificationOnCallWindow) []oncall.Spec {
	specs := make([]oncall.Spec, 0, len(windows))
	for _, window := range windows {
		specs = append(specs, oncall.Spec{Weekdays: window.Weekdays, Start: window.Start, End: window.End})
	}
	return specs
}

// onCallWindows 把规范化后的时段转回业务对象。
func onCallWindows(specs []oncall.Spec) []bo.NotificationOnCallWindow {
	windows := make([]bo.NotificationOnCallWindow, 0, len(specs))
	for _, spec := range specs {
		windows = append(windows, bo.NotificationOnCallWindow{
			Weekdays: spec.Weekdays,
			Start:    spec.Start,
			End:      spec.End,
		})
	}
	return windows
}
