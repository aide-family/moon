// Package probe 管理探测目标组：校验、归一化、试探测与立即探测，
// 以及在目标被删除或停用时收口旧事件。探测规则的判定在 alertevent 里。
package probe

import (
	"context"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	labelsutil "github.com/aide-family/moon/pkg/labels"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/probe"
)

var ProviderSet = wire.NewSet(NewGateway)

const (
	// maxGroupName 与 proto 校验保持一致。
	maxGroupName = 64
	// maxTargets 是单个目标组的地址上限，和 proto 的 repeated.max_items 一致。
	maxTargets = 50
	// maxRemark 是备注长度上限。
	maxRemark = 200
)

// Gateway 管理探测目标组。
type Gateway struct {
	targets repository.ProbeTargetRepo
	events  repository.AlertEventRepo
	access  access.Checker
	log     *slog.Logger
}

func NewGateway(targets repository.ProbeTargetRepo, events repository.AlertEventRepo, checker access.Checker, logger *slog.Logger) *Gateway {
	return &Gateway{targets: targets, events: events, access: checker, log: logger}
}

// CreateProbeTargetGroup 校验并保存目标组。
func (gateway *Gateway) CreateProbeTargetGroup(ctx context.Context, in bo.ProbeTargetGroup) (*bo.ProbeTargetGroup, error) {
	ns, err := gateway.access.Allow(ctx, perm.ProbeWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := normalizeGroup(&in); err != nil {
		return nil, err
	}
	if err := gateway.requireDriverCapability(ctx, in.Driver); err != nil {
		return nil, err
	}
	item, err := gateway.targets.CreateProbeTargetGroup(ctx, in)
	return item, gateway.access.Wrap(err)
}

// UpdateProbeTargetGroup 整体替换目标列表，并把被移除或停用的目标上的旧事件收口。
func (gateway *Gateway) UpdateProbeTargetGroup(ctx context.Context, in bo.ProbeTargetGroup) (*bo.ProbeTargetGroup, error) {
	ns, err := gateway.access.Allow(ctx, perm.ProbeWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := normalizeGroup(&in); err != nil {
		return nil, err
	}
	previous, err := gateway.targets.GetProbeTargetGroup(ctx, ns, in.UID)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	// 新驱动和原驱动都必须已开通：未授权类型的目标组不能通过已知 UID 修改。
	if err := gateway.requireDriverCapability(ctx, previous.Driver); err != nil {
		return nil, err
	}
	item, err := gateway.targets.UpdateProbeTargetGroup(ctx, in)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	gateway.recoverRemoved(ctx, ns, previous, item)
	return item, nil
}

// UpdateProbeTargetGroupStatus 启停目标组。停用后它的目标上的未恢复事件立即收口。
func (gateway *Gateway) UpdateProbeTargetGroupStatus(ctx context.Context, uid, status string) (*bo.ProbeTargetGroup, error) {
	ns, err := gateway.access.Allow(ctx, perm.ProbeWrite)
	if err != nil {
		return nil, err
	}
	if status != "ACTIVE" && status != "DISABLED" {
		return nil, merr.Invalid("invalid status")
	}
	current, err := gateway.targets.GetProbeTargetGroup(ctx, ns, uid)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	if err := gateway.requireDriverCapability(ctx, current.Driver); err != nil {
		return nil, err
	}
	item, err := gateway.targets.UpdateProbeTargetGroupStatus(ctx, ns, uid, status)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	if status == "DISABLED" {
		gateway.recoverTargets(ctx, ns, item.Targets)
	}
	return item, nil
}

// DeleteProbeTargetGroup 删除目标组并收口它的全部目标事件。
func (gateway *Gateway) DeleteProbeTargetGroup(ctx context.Context, uid string) error {
	ns, err := gateway.access.Allow(ctx, perm.ProbeDelete)
	if err != nil {
		return err
	}
	item, err := gateway.targets.GetProbeTargetGroup(ctx, ns, uid)
	if err != nil {
		return gateway.access.Wrap(err)
	}
	if err := gateway.requireDriverCapability(ctx, item.Driver); err != nil {
		return err
	}
	if err := gateway.targets.DeleteProbeTargetGroup(ctx, ns, uid); err != nil {
		return gateway.access.Wrap(err)
	}
	gateway.recoverTargets(ctx, ns, item.Targets)
	return nil
}

// GetProbeTargetGroup 返回目标组详情。
func (gateway *Gateway) GetProbeTargetGroup(ctx context.Context, uid string) (*bo.ProbeTargetGroup, error) {
	ns, err := gateway.access.Allow(ctx, perm.ProbeRead)
	if err != nil {
		return nil, err
	}
	item, err := gateway.targets.GetProbeTargetGroup(ctx, ns, uid)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	if err := gateway.requireDriverCapability(ctx, item.Driver); err != nil {
		return nil, err
	}
	return item, nil
}

// ListProbeTargetGroups 分页列出目标组。
func (gateway *Gateway) ListProbeTargetGroups(ctx context.Context, pageNo, pageSize int32, name, driver string) ([]bo.ProbeTargetGroup, int, int, int, error) {
	ns, err := gateway.access.Allow(ctx, perm.ProbeRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	drivers, err := gateway.authorizedDrivers(ctx)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	if requested := strings.TrimSpace(driver); requested != "" {
		drivers = filterDriver(drivers, requested)
	}
	if len(drivers) == 0 {
		return []bo.ProbeTargetGroup{}, 0, p, s, nil
	}
	items, total, err := gateway.targets.ListProbeTargetGroups(ctx, ns, p, s, strings.TrimSpace(name), drivers)
	return items, total, p, s, gateway.access.Wrap(err)
}

// requireDriverCapability 要求当前租户已开通该探测驱动对应的能力。
func (gateway *Gateway) requireDriverCapability(ctx context.Context, driver string) error {
	capability := perm.ProbeCapabilityOf(driver)
	if capability == "" {
		return merr.Invalid("invalid request")
	}
	return gateway.access.RequireCapability(ctx, capability)
}

// authorizedDrivers 返回当前租户已开通的探测驱动。
func (gateway *Gateway) authorizedDrivers(ctx context.Context) ([]string, error) {
	capabilities, err := gateway.access.Capabilities(ctx)
	if err != nil {
		return nil, err
	}
	return perm.ProbeDriversOf(capabilities), nil
}

// filterDriver 只保留请求的类型；类型未开通时返回空，列表不会退回全量。
func filterDriver(drivers []string, wanted string) []string {
	for _, driver := range drivers {
		if driver == wanted {
			return []string{wanted}
		}
	}
	return nil
}

// ListProbeTargetResults 分页返回某个目标组的探测记录。
func (gateway *Gateway) ListProbeTargetResults(ctx context.Context, groupUID, targetUID string, pageNo, pageSize int32) ([]bo.ProbeResult, int, int, int, error) {
	ns, err := gateway.access.Allow(ctx, perm.ProbeRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	group, err := gateway.targets.GetProbeTargetGroup(ctx, ns, groupUID)
	if err != nil {
		return nil, 0, 0, 0, gateway.access.Wrap(err)
	}
	if err := gateway.requireDriverCapability(ctx, group.Driver); err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := gateway.targets.ListProbeTargetResults(ctx, ns, groupUID, strings.TrimSpace(targetUID), p, s)
	return items, total, p, s, gateway.access.Wrap(err)
}

// ProbeTargetGroupNow 立即探测一次组内全部启用地址：写探测记录，但不按规则级别判定。
func (gateway *Gateway) ProbeTargetGroupNow(ctx context.Context, uid string) ([]bo.ProbeResult, error) {
	ns, err := gateway.access.Allow(ctx, perm.ProbeWrite)
	if err != nil {
		return nil, err
	}
	group, err := gateway.targets.GetProbeTargetGroup(ctx, ns, uid)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	if err := gateway.requireDriverCapability(ctx, group.Driver); err != nil {
		return nil, err
	}
	request, err := probe.ParseRequest(group.Request)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	slot := now.Truncate(time.Second)
	results := make([]bo.ProbeResult, 0, len(group.Targets))
	for _, target := range group.Targets {
		if target.Status != "ACTIVE" {
			continue
		}
		runCtx, cancel := context.WithTimeout(ctx, time.Duration(request.TimeoutMS)*time.Millisecond)
		outcome := probe.Run(runCtx, target.Spec, request)
		cancel()
		results = append(results, toResult(ns, group.UID, target, outcome, slot, now))
	}
	if _, err := gateway.targets.SaveProbeResults(ctx, results); err != nil {
		return nil, gateway.access.Wrap(err)
	}
	return results, nil
}

// ProbeTrial 对单个地址试探测：不落库、不判定、不产生事件。
func (gateway *Gateway) ProbeTrial(ctx context.Context, spec string, request probe.Request) (*bo.ProbeResult, error) {
	if _, err := gateway.access.Allow(ctx, perm.ProbeWrite); err != nil {
		return nil, err
	}
	if err := gateway.requireDriverCapability(ctx, bo.ProbeDriverHTTP); err != nil {
		return nil, err
	}
	normalized, err := probe.NormalizeSpec(spec)
	if err != nil {
		return nil, err
	}
	request = request.Normalize()
	if err := request.Validate(); err != nil {
		return nil, err
	}
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(request.TimeoutMS)*time.Millisecond)
	defer cancel()
	outcome := probe.Run(runCtx, normalized, request)
	now := time.Now().UTC()
	result := toResult("", "", bo.ProbeTarget{Spec: normalized}, outcome, now.Truncate(time.Second), now)
	return &result, nil
}

// normalizeGroup 校验并归一化名称、标签、请求参数与目标列表。
func normalizeGroup(in *bo.ProbeTargetGroup) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return merr.Invalid("name is required")
	}
	if utf8.RuneCountInString(in.Name) > maxGroupName {
		return merr.Invalid("name is too long")
	}
	if utf8.RuneCountInString(in.Remark) > maxRemark {
		return merr.Invalid("remark is too long")
	}
	if in.Driver == "" {
		in.Driver = bo.ProbeDriverHTTP
	}
	if in.Driver != bo.ProbeDriverHTTP {
		return merr.Invalid("unsupported probe driver")
	}
	labels, err := labelsutil.NormalizeObject(in.Labels)
	if err != nil {
		return err
	}
	in.Labels = labels
	executorLabels, err := labelsutil.NormalizeObject(in.ExecutorLabels)
	if err != nil {
		return err
	}
	if executorLabels == "" {
		executorLabels = "{}"
	}
	in.ExecutorLabels = executorLabels

	request, err := probe.ParseRequest(in.Request)
	if err != nil {
		return err
	}
	request = request.Normalize()
	if err := request.Validate(); err != nil {
		return err
	}
	in.Request = request.Marshal()

	if len(in.Targets) == 0 {
		return merr.Invalid("at least one probe target is required")
	}
	if len(in.Targets) > maxTargets {
		return merr.Invalid("too many probe targets")
	}
	seen := make(map[string]struct{}, len(in.Targets))
	for index := range in.Targets {
		target := &in.Targets[index]
		spec, err := probe.NormalizeSpec(target.Spec)
		if err != nil {
			return err
		}
		if _, ok := seen[spec]; ok {
			return merr.Invalid("duplicate probe target")
		}
		seen[spec] = struct{}{}
		target.Spec = spec
		if target.Status == "" {
			target.Status = "ACTIVE"
		}
		if target.Status != "ACTIVE" && target.Status != "DISABLED" {
			return merr.Invalid("invalid target status")
		}
		if target.Labels == "" {
			target.Labels = "{}"
		}
		normalized, err := labelsutil.NormalizeObject(target.Labels)
		if err != nil {
			return err
		}
		target.Labels = normalized
		target.Sort = index
	}
	return nil
}

// recoverRemoved 收口在更新中被移除或停用的目标。
func (gateway *Gateway) recoverRemoved(ctx context.Context, namespaceUID string, previous, current *bo.ProbeTargetGroup) {
	if previous == nil || current == nil {
		return
	}
	alive := make(map[string]string, len(current.Targets))
	for _, target := range current.Targets {
		alive[target.UID] = target.Status
	}
	removed := make([]bo.ProbeTarget, 0)
	for _, target := range previous.Targets {
		status, ok := alive[target.UID]
		if !ok || status != "ACTIVE" {
			removed = append(removed, target)
		}
	}
	gateway.recoverTargets(ctx, namespaceUID, removed)
}

// recoverTargets 收口一批目标上的未恢复事件；失败只记日志，不影响主流程。
func (gateway *Gateway) recoverTargets(ctx context.Context, namespaceUID string, targets []bo.ProbeTarget) {
	if len(targets) == 0 {
		return
	}
	uids := make([]string, 0, len(targets))
	for _, target := range targets {
		if target.UID != "" {
			uids = append(uids, target.UID)
		}
	}
	if len(uids) == 0 {
		return
	}
	if err := gateway.events.RecoverProbeTargets(ctx, namespaceUID, uids); err != nil {
		gateway.log.Error("recover probe events: " + err.Error())
	}
}

// toResult 把一次执行结果转成入库结构。
func toResult(namespaceUID, groupUID string, target bo.ProbeTarget, outcome probe.Result, slot, at time.Time) bo.ProbeResult {
	value := "0"
	if outcome.OK {
		value = "1"
	}
	return bo.ProbeResult{
		NamespaceUID: namespaceUID,
		GroupUID:     groupUID,
		TargetUID:    target.UID,
		TargetSpec:   target.Spec,
		Slot:         slot,
		ProbedAt:     at,
		OK:           outcome.OK,
		Value:        value,
		DurationMS:   int(outcome.DurationMS),
		StatusCode:   outcome.StatusCode,
		ErrorKind:    outcome.ErrorKind,
		ErrorMessage: outcome.ErrorMessage,
		DNSMS:        int(outcome.DNSMS),
		TCPMS:        int(outcome.TCPMS),
		TLSMS:        int(outcome.TLSMS),
		TTFBMS:       int(outcome.TTFBMS),
		CertDays:     outcome.CertDays,
		ResolvedIP:   outcome.ResolvedIP,
	}
}
