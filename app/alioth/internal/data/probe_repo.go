package data

import (
	"context"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertrule"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/proberesult"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/probetarget"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/probetargetgroup"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	labelsutil "github.com/aide-family/moon/pkg/labels"
)

func (r *probeRepo) CreateProbeTargetGroup(ctx context.Context, in bo.ProbeTargetGroup) (*bo.ProbeTargetGroup, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	request, err := encryptProbeRequest(r.box, in.Request)
	if err != nil {
		return nil, err
	}
	row, err := tx.ProbeTargetGroup.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetDriver(probeGroupDriver(in.Driver)).
		SetStatus(probetargetgroup.StatusACTIVE).
		SetLabels(in.Labels).
		SetExecutorLabels(in.ExecutorLabels).
		SetRequest(request).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	for _, target := range in.Targets {
		if _, err := tx.ProbeTarget.Create().
			SetUID(id.New()).
			SetNamespaceUID(in.NamespaceUID).
			SetGroupUID(row.UID).
			SetSpec(target.Spec).
			SetLabels(target.Labels).
			SetStatus(probetarget.Status(target.Status)).
			SetSort(target.Sort).
			Save(ctx); err != nil {
			return nil, r.fail(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetProbeTargetGroup(ctx, in.NamespaceUID, row.UID)
}

// UpdateProbeTargetGroup 整体替换目标列表：带 UID 的更新，带过的保留，缺少的删除。
func (r *probeRepo) UpdateProbeTargetGroup(ctx context.Context, in bo.ProbeTargetGroup) (*bo.ProbeTargetGroup, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	request, err := encryptProbeRequest(r.box, in.Request)
	if err != nil {
		return nil, err
	}
	if _, err := tx.ProbeTargetGroup.Update().Where(
		probetargetgroup.NamespaceUIDEQ(in.NamespaceUID),
		probetargetgroup.UIDEQ(in.UID),
		probetargetgroup.DeletedAtIsNil(),
	).
		SetName(in.Name).
		SetDriver(probeGroupDriver(in.Driver)).
		SetLabels(in.Labels).
		SetExecutorLabels(in.ExecutorLabels).
		SetRequest(request).
		SetRemark(in.Remark).
		Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	keep := make(map[string]struct{}, len(in.Targets))
	for _, target := range in.Targets {
		if target.UID == "" {
			created, err := tx.ProbeTarget.Create().
				SetUID(id.New()).
				SetNamespaceUID(in.NamespaceUID).
				SetGroupUID(in.UID).
				SetSpec(target.Spec).
				SetLabels(target.Labels).
				SetStatus(probetarget.Status(target.Status)).
				SetSort(target.Sort).
				Save(ctx)
			if err != nil {
				return nil, r.fail(err)
			}
			// 新建的目标也要进 keep，否则会被下面的「不在 payload 里的旧行」清扫误删。
			keep[created.UID] = struct{}{}
			continue
		}
		keep[target.UID] = struct{}{}
		if _, err := tx.ProbeTarget.Update().Where(
			probetarget.NamespaceUIDEQ(in.NamespaceUID),
			probetarget.GroupUIDEQ(in.UID),
			probetarget.UIDEQ(target.UID),
			probetarget.DeletedAtIsNil(),
		).
			SetSpec(target.Spec).
			SetLabels(target.Labels).
			SetStatus(probetarget.Status(target.Status)).
			SetSort(target.Sort).
			Save(ctx); err != nil {
			return nil, r.fail(err)
		}
	}
	query := tx.ProbeTarget.Query().Where(
		probetarget.NamespaceUIDEQ(in.NamespaceUID),
		probetarget.GroupUIDEQ(in.UID),
		probetarget.DeletedAtIsNil(),
	)
	rows, err := query.All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	for _, row := range rows {
		if _, ok := keep[row.UID]; ok {
			continue
		}
		if err := tx.ProbeTarget.Update().Where(probetarget.UIDEQ(row.UID)).SetDeletedAt(time.Now()).Exec(ctx); err != nil {
			return nil, r.fail(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetProbeTargetGroup(ctx, in.NamespaceUID, in.UID)
}

func (r *probeRepo) UpdateProbeTargetGroupStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.ProbeTargetGroup, error) {
	if _, err := r.client.ProbeTargetGroup.Update().Where(
		probetargetgroup.NamespaceUIDEQ(namespaceUID),
		probetargetgroup.UIDEQ(uid),
		probetargetgroup.DeletedAtIsNil(),
	).SetStatus(probetargetgroup.Status(status)).Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	return r.GetProbeTargetGroup(ctx, namespaceUID, uid)
}

func (r *probeRepo) DeleteProbeTargetGroup(ctx context.Context, namespaceUID, uid string) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now()
	if err := tx.ProbeTargetGroup.Update().Where(
		probetargetgroup.NamespaceUIDEQ(namespaceUID),
		probetargetgroup.UIDEQ(uid),
		probetargetgroup.DeletedAtIsNil(),
	).SetDeletedAt(now).Exec(ctx); err != nil {
		return r.fail(err)
	}
	if err := tx.ProbeTarget.Update().Where(
		probetarget.NamespaceUIDEQ(namespaceUID),
		probetarget.GroupUIDEQ(uid),
	).SetDeletedAt(now).Exec(ctx); err != nil {
		return r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *probeRepo) GetProbeTargetGroup(ctx context.Context, namespaceUID, uid string) (*bo.ProbeTargetGroup, error) {
	row, err := r.client.ProbeTargetGroup.Query().Where(
		probetargetgroup.NamespaceUIDEQ(namespaceUID),
		probetargetgroup.UIDEQ(uid),
		probetargetgroup.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item, err := r.toProbeGroup(row)
	if err != nil {
		return nil, err
	}
	if err := r.fillProbeTargets(ctx, []*bo.ProbeTargetGroup{&item}); err != nil {
		return nil, err
	}
	r.fillProbeLatest(ctx, []*bo.ProbeTargetGroup{&item})
	return &item, nil
}

func (r *probeRepo) ListProbeTargetGroups(ctx context.Context, namespaceUID string, page, pageSize int, name string, drivers []string) ([]bo.ProbeTargetGroup, int, error) {
	query := r.client.ProbeTargetGroup.Query().Where(
		probetargetgroup.NamespaceUIDEQ(namespaceUID),
		probetargetgroup.DeletedAtIsNil(),
	)
	if name != "" {
		query = query.Where(probetargetgroup.NameContainsFold(name))
	}
	if len(drivers) > 0 {
		values := make([]probetargetgroup.Driver, 0, len(drivers))
		for _, driver := range drivers {
			values = append(values, probetargetgroup.Driver(driver))
		}
		query = query.Where(probetargetgroup.DriverIn(values...))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(probetargetgroup.FieldName), ent.Asc(probetargetgroup.FieldUID)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.ProbeTargetGroup, 0, len(rows))
	for _, row := range rows {
		item, err := r.toProbeGroup(row)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := r.fillProbeTargets(ctx, probeGroupsRef(items)); err != nil {
		return nil, 0, err
	}
	r.fillProbeLatest(ctx, probeGroupsRef(items))
	return items, total, nil
}

func (r *probeRepo) ListProbeTargetGroupsByMatchers(ctx context.Context, namespaceUID, driver, matchersJSON string) ([]bo.ProbeTargetGroup, error) {
	query := r.client.ProbeTargetGroup.Query().Where(
		probetargetgroup.NamespaceUIDEQ(namespaceUID),
		probetargetgroup.StatusEQ(probetargetgroup.StatusACTIVE),
		probetargetgroup.DeletedAtIsNil(),
	)
	if driver != "" {
		query = query.Where(probetargetgroup.DriverEQ(probetargetgroup.Driver(driver)))
	}
	rows, err := query.Order(ent.Asc(probetargetgroup.FieldName)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	matchers, err := labelsutil.ParseObject(matchersJSON)
	if err != nil {
		return nil, err
	}
	match := labelsutil.CompileMatchers(matchers)
	items := make([]bo.ProbeTargetGroup, 0, len(rows))
	for _, row := range rows {
		if !match(probeGroupMatchLabels(row)) {
			continue
		}
		item, err := r.toProbeGroup(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil, nil
	}
	if err := r.fillProbeTargets(ctx, probeGroupsRef(items)); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *probeRepo) ListProbeTargetGroupsByExecutor(ctx context.Context, executorKind, executorLabels string) ([]bo.ProbeTargetGroup, error) {
	query := r.client.ProbeTargetGroup.Query().Where(
		probetargetgroup.StatusEQ(probetargetgroup.StatusACTIVE),
		probetargetgroup.DeletedAtIsNil(),
	)
	if executorKind == "CLUSTER" {
		// 集群内执行器只领 executor_labels 为空的目标组。
		query = query.Where(probetargetgroup.ExecutorLabelsEQ("{}"))
	}
	rows, err := query.Order(ent.Asc(probetargetgroup.FieldName)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.ProbeTargetGroup, 0, len(rows))
	for _, row := range rows {
		item, err := r.toProbeGroup(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil, nil
	}
	if err := r.fillProbeTargets(ctx, probeGroupsRef(items)); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *probeRepo) GetProbeTarget(ctx context.Context, namespaceUID, uid string) (*bo.ProbeTarget, error) {
	row, err := r.client.ProbeTarget.Query().Where(
		probetarget.NamespaceUIDEQ(namespaceUID),
		probetarget.UIDEQ(uid),
		probetarget.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toProbeTarget(row)
	return &item, nil
}

func (r *probeRepo) ListProbeTargets(ctx context.Context, namespaceUID, groupUID string) ([]bo.ProbeTarget, error) {
	rows, err := r.client.ProbeTarget.Query().Where(
		probetarget.NamespaceUIDEQ(namespaceUID),
		probetarget.GroupUIDEQ(groupUID),
		probetarget.StatusEQ(probetarget.StatusACTIVE),
		probetarget.DeletedAtIsNil(),
	).Order(ent.Asc(probetarget.FieldSort), ent.Asc(probetarget.FieldUID)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	out := make([]bo.ProbeTarget, 0, len(rows))
	for _, row := range rows {
		out = append(out, toProbeTarget(row))
	}
	return out, nil
}

// fillProbeTargets 一次补齐多组的目标与计数。
func (r *probeRepo) fillProbeTargets(ctx context.Context, groups []*bo.ProbeTargetGroup) error {
	if len(groups) == 0 {
		return nil
	}
	uids := make([]string, 0, len(groups))
	for _, group := range groups {
		uids = append(uids, group.UID)
	}
	rows, err := r.client.ProbeTarget.Query().Where(
		probetarget.GroupUIDIn(uids...),
		probetarget.DeletedAtIsNil(),
	).Order(ent.Asc(probetarget.FieldSort), ent.Asc(probetarget.FieldUID)).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	byGroup := make(map[string][]bo.ProbeTarget, len(groups))
	for _, row := range rows {
		byGroup[row.GroupUID] = append(byGroup[row.GroupUID], toProbeTarget(row))
	}
	for _, group := range groups {
		targets := byGroup[group.UID]
		group.Targets = targets
		group.TargetCount = len(targets)
		active := 0
		for _, target := range targets {
			if target.Status == "ACTIVE" {
				active++
			}
		}
		group.ActiveTargetCount = active
	}
	return nil
}

// fillProbeLatest 补齐每组最新一条探测结果与最近探测时间。
func (r *probeRepo) fillProbeLatest(ctx context.Context, groups []*bo.ProbeTargetGroup) {
	targetUIDs := make([]string, 0)
	for _, group := range groups {
		for _, target := range group.Targets {
			targetUIDs = append(targetUIDs, target.UID)
		}
	}
	latest, err := r.LatestProbeResults(ctx, targetUIDs)
	if err != nil {
		return
	}
	for _, group := range groups {
		var newest *bo.ProbeResult
		for index := range group.Targets {
			if result, ok := latest[group.Targets[index].UID]; ok {
				copied := result
				group.Targets[index].LastResult = &copied
				if newest == nil || copied.ProbedAt.After(newest.ProbedAt) {
					newest = &copied
				}
			}
		}
		if newest != nil {
			at := newest.ProbedAt
			group.LastProbedAt = &at
		}
	}
	_ = targetUIDs
}

func (r *probeRepo) SaveProbeResults(ctx context.Context, results []bo.ProbeResult) (map[string]struct{}, error) {
	inserted := make(map[string]struct{}, len(results))
	if len(results) == 0 {
		return inserted, nil
	}
	for _, item := range results {
		_, err := r.client.ProbeResult.Create().
			SetUID(id.New()).
			SetNamespaceUID(item.NamespaceUID).
			SetRuleUID(item.RuleUID).
			SetGroupUID(item.GroupUID).
			SetTargetUID(item.TargetUID).
			SetSlot(item.Slot).
			SetProbedAt(item.ProbedAt).
			SetOk(item.OK).
			SetValue(item.Value).
			SetDurationMs(item.DurationMS).
			SetStatusCode(item.StatusCode).
			SetErrorKind(item.ErrorKind).
			SetErrorMessage(item.ErrorMessage).
			SetDNSMs(item.DNSMS).
			SetTCPMs(item.TCPMS).
			SetTLSMs(item.TLSMS).
			SetTtfbMs(item.TTFBMS).
			SetCertDays(item.CertDays).
			SetResolvedIP(item.ResolvedIP).
			SetExecutor(item.Executor).
			SetPacketLoss(item.PacketLoss).
			SetJitterMs(item.JitterMS).
			SetMinRttMs(item.MinRTTMS).
			SetMaxRttMs(item.MaxRTTMS).
			SetAnswerCount(item.AnswerCount).
			SetDNSAnswers(item.DNSAnswers).
			SetTLSVersion(item.TLSVersion).
			SetTLSCipher(item.TLSCipher).
			SetCertSubject(item.CertSubject).
			SetCertIssuer(item.CertIssuer).
			SetCertSans(item.CertSANS).
			Save(ctx)
		if err != nil {
			if ent.IsConstraintError(err) {
				// 同一 (target, slot) 已回写过，幂等跳过。
				continue
			}
			return inserted, r.fail(err)
		}
		inserted[item.TargetUID] = struct{}{}
	}
	return inserted, nil
}

// ProbingTargetUIDs 返回某个槽位已有结果的目标；领取时跳过它们，避免同一个
// 槽位在窗口内被反复探测和反复回写。
func (r *probeRepo) ProbingTargetUIDs(ctx context.Context, namespaceUID string, slot time.Time, targetUIDs []string) (map[string]struct{}, error) {
	out := make(map[string]struct{}, len(targetUIDs))
	if len(targetUIDs) == 0 {
		return out, nil
	}
	rows, err := r.client.ProbeResult.Query().Where(
		proberesult.NamespaceUIDEQ(namespaceUID),
		proberesult.SlotEQ(slot),
		proberesult.TargetUIDIn(targetUIDs...),
	).Select(proberesult.FieldTargetUID).All(ctx)
	if err != nil {
		return out, r.fail(err)
	}
	for _, row := range rows {
		out[row.TargetUID] = struct{}{}
	}
	return out, nil
}

func (r *probeRepo) ListProbeResults(ctx context.Context, namespaceUID string, targetUIDs []string, from, to time.Time) ([]bo.ProbeResult, error) {
	if len(targetUIDs) == 0 {
		return nil, nil
	}
	rows, err := r.client.ProbeResult.Query().Where(
		proberesult.NamespaceUIDEQ(namespaceUID),
		proberesult.TargetUIDIn(targetUIDs...),
		proberesult.SlotGTE(from),
		proberesult.SlotLTE(to),
	).Order(ent.Asc(proberesult.FieldSlot)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	out := make([]bo.ProbeResult, 0, len(rows))
	for _, row := range rows {
		out = append(out, toProbeResult(row))
	}
	return out, nil
}

func (r *probeRepo) ListProbeTargetResults(ctx context.Context, namespaceUID, groupUID, targetUID string, page, pageSize int) ([]bo.ProbeResult, int, error) {
	query := r.client.ProbeResult.Query().Where(
		proberesult.NamespaceUIDEQ(namespaceUID),
		proberesult.GroupUIDEQ(groupUID),
	)
	if targetUID != "" {
		query = query.Where(proberesult.TargetUIDEQ(targetUID))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(proberesult.FieldProbedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	out := make([]bo.ProbeResult, 0, len(rows))
	for _, row := range rows {
		out = append(out, toProbeResult(row))
	}
	r.fillRuleNames(ctx, out)
	r.fillTargetSpecs(ctx, out)
	return out, total, nil
}

func (r *probeRepo) LatestProbeResults(ctx context.Context, targetUIDs []string) (map[string]bo.ProbeResult, error) {
	out := make(map[string]bo.ProbeResult, len(targetUIDs))
	if len(targetUIDs) == 0 {
		return out, nil
	}
	rows, err := r.client.ProbeResult.Query().Where(
		proberesult.TargetUIDIn(targetUIDs...),
	).Order(ent.Desc(proberesult.FieldProbedAt)).All(ctx)
	if err != nil {
		return out, r.fail(err)
	}
	if len(rows) == 0 {
		return out, nil
	}
	specs := r.probeTargetSpecs(ctx, targetUIDs)
	for _, row := range rows {
		if _, ok := out[row.TargetUID]; ok {
			continue
		}
		item := toProbeResult(row)
		item.TargetSpec = specs[row.TargetUID]
		out[row.TargetUID] = item
	}
	return out, nil
}

// fillTargetSpecs 按 target_uid 批量补探测记录的目标地址。目标行可能已经被软删，
// 但历史记录仍要显示地址，所以查询故意不过滤 deleted_at；目标行确实不存在时留空。
func (r *probeRepo) fillTargetSpecs(ctx context.Context, results []bo.ProbeResult) {
	if len(results) == 0 {
		return
	}
	seen := make(map[string]struct{}, len(results))
	uids := make([]string, 0, len(results))
	for _, item := range results {
		if item.TargetUID == "" {
			continue
		}
		if _, ok := seen[item.TargetUID]; ok {
			continue
		}
		seen[item.TargetUID] = struct{}{}
		uids = append(uids, item.TargetUID)
	}
	specs := r.probeTargetSpecs(ctx, uids)
	for index := range results {
		results[index].TargetSpec = specs[results[index].TargetUID]
	}
}

// probeTargetSpecs 返回 target_uid -> spec。不过滤 deleted_at，保证软删目标的
// 历史探测记录也能显示地址。
func (r *probeRepo) probeTargetSpecs(ctx context.Context, targetUIDs []string) map[string]string {
	specs := make(map[string]string, len(targetUIDs))
	if len(targetUIDs) == 0 {
		return specs
	}
	rows, err := r.client.ProbeTarget.Query().
		Where(probetarget.UIDIn(targetUIDs...)).
		Select(probetarget.FieldUID, probetarget.FieldSpec).
		All(ctx)
	if err != nil {
		return specs
	}
	for _, row := range rows {
		specs[row.UID] = row.Spec
	}
	return specs
}

// fillRuleNames 给探测记录补上规则名称。目标组可以被多条规则匹配，记录按规则产生，
// 列表里需要显示是哪条规则判定的。
func (r *probeRepo) fillRuleNames(ctx context.Context, results []bo.ProbeResult) {
	if len(results) == 0 {
		return
	}
	seen := make(map[string]struct{}, len(results))
	uids := make([]string, 0, len(results))
	for _, item := range results {
		if item.RuleUID == "" {
			continue
		}
		if _, ok := seen[item.RuleUID]; ok {
			continue
		}
		seen[item.RuleUID] = struct{}{}
		uids = append(uids, item.RuleUID)
	}
	if len(uids) == 0 {
		return
	}
	rows, err := r.client.AlertRule.Query().Where(alertrule.UIDIn(uids...)).All(ctx)
	if err != nil {
		return
	}
	names := make(map[string]string, len(rows))
	for _, row := range rows {
		names[row.UID] = row.Name
	}
	for index := range results {
		results[index].RuleName = names[results[index].RuleUID]
	}
}

func (r *probeRepo) PruneProbeResults(ctx context.Context, before time.Time, keepPerTarget int) (int, error) {
	deleted, err := r.client.ProbeResult.Delete().Where(proberesult.ProbedAtLT(before)).Exec(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	return deleted, nil
}

// probeGroupsRef 返回可写指针，便于 fill 系列函数原地补齐列表项。
func probeGroupsRef(items []bo.ProbeTargetGroup) []*bo.ProbeTargetGroup {
	out := make([]*bo.ProbeTargetGroup, 0, len(items))
	for i := range items {
		out = append(out, &items[i])
	}
	return out
}

// probeGroupDriver 把业务层归一化后的驱动落库。驱动由 biz 校验；空值兜底
// PROBE_HTTP，兼容只传目标列表、不关心类型的直接仓储调用与旧数据。
func probeGroupDriver(driver string) probetargetgroup.Driver {
	if driver == "" {
		return probetargetgroup.DriverPROBE_HTTP
	}
	return probetargetgroup.Driver(driver)
}

// probeRequestCipherPrefix 标记加密后的目标组请求 JSON，读取时据此区分历史明文行。
const probeRequestCipherPrefix = "enc:v1:"

// encryptProbeRequest 加密目标组请求 JSON（含自定义请求头里的 token），密文
// 带 enc:v1: 前缀。已经带前缀的值原样返回，避免把密文再加密一次。
func encryptProbeRequest(box *secretBox, plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	if strings.HasPrefix(plain, probeRequestCipherPrefix) {
		return plain, nil
	}
	cipher, err := box.encrypt(plain)
	if err != nil {
		return "", err
	}
	return probeRequestCipherPrefix + cipher, nil
}

// decryptProbeRequest 解密目标组请求 JSON。没有前缀按历史明文行原样返回，
// 该行在下一次保存时自动加密；有前缀但解密失败（密钥不一致或数据损坏）
// 返回服务端错误，不能静默当成空请求。
func decryptProbeRequest(box *secretBox, stored string) (string, error) {
	if stored == "" {
		return "", nil
	}
	if !strings.HasPrefix(stored, probeRequestCipherPrefix) {
		return stored, nil
	}
	return box.decrypt(strings.TrimPrefix(stored, probeRequestCipherPrefix))
}

// toProbeGroup 把目标组行转成业务对象，并把加密的请求 JSON 解密回明文。
func (r *probeRepo) toProbeGroup(row *ent.ProbeTargetGroup) (bo.ProbeTargetGroup, error) {
	request, err := decryptProbeRequest(r.box, row.Request)
	if err != nil {
		return bo.ProbeTargetGroup{}, err
	}
	return bo.ProbeTargetGroup{
		UID:            row.UID,
		NamespaceUID:   row.NamespaceUID,
		Name:           row.Name,
		Driver:         string(row.Driver),
		Status:         string(row.Status),
		Labels:         row.Labels,
		ExecutorLabels: row.ExecutorLabels,
		Request:        request,
		Remark:         row.Remark,
	}, nil
}

func toProbeTarget(row *ent.ProbeTarget) bo.ProbeTarget {
	return bo.ProbeTarget{
		UID:          row.UID,
		NamespaceUID: row.NamespaceUID,
		GroupUID:     row.GroupUID,
		Spec:         row.Spec,
		Labels:       row.Labels,
		Status:       string(row.Status),
		Sort:         row.Sort,
	}
}

func toProbeResult(row *ent.ProbeResult) bo.ProbeResult {
	return bo.ProbeResult{
		UID:          row.UID,
		NamespaceUID: row.NamespaceUID,
		RuleUID:      row.RuleUID,
		GroupUID:     row.GroupUID,
		TargetUID:    row.TargetUID,
		Slot:         row.Slot,
		ProbedAt:     row.ProbedAt,
		OK:           row.Ok,
		Value:        row.Value,
		DurationMS:   row.DurationMs,
		StatusCode:   row.StatusCode,
		ErrorKind:    row.ErrorKind,
		ErrorMessage: row.ErrorMessage,
		DNSMS:        row.DNSMs,
		TCPMS:        row.TCPMs,
		TLSMS:        row.TLSMs,
		TTFBMS:       row.TtfbMs,
		CertDays:     row.CertDays,
		ResolvedIP:   row.ResolvedIP,
		Executor:     row.Executor,
		PacketLoss:   row.PacketLoss,
		JitterMS:     row.JitterMs,
		MinRTTMS:     row.MinRttMs,
		MaxRTTMS:     row.MaxRttMs,
		AnswerCount:  row.AnswerCount,
		DNSAnswers:   row.DNSAnswers,
		TLSVersion:   row.TLSVersion,
		TLSCipher:    row.TLSCipher,
		CertSubject:  row.CertSubject,
		CertIssuer:   row.CertIssuer,
		CertSANS:     row.CertSans,
	}
}

// probeGroupMatchLabels 把目标组的匹配键摊平：固定键 probe_group 是组名，
// 其余键来自目标组 labels，语义与数据源标签匹配一致。
func probeGroupMatchLabels(row *ent.ProbeTargetGroup) map[string]string {
	labels := map[string]string{"probe_group": row.Name}
	parsed, err := labelsutil.ParseObject(row.Labels)
	if err != nil {
		return labels
	}
	for key, value := range parsed {
		if _, exists := labels[key]; !exists {
			labels[key] = value
		}
	}
	return labels
}
