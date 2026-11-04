package oncallrotation

import (
	"context"
	"log/slog"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/oncall"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewGateway)

const (
	// maxNameRunes 与 proto 校验保持一致。
	maxNameRunes = 64
	// maxRemarkRunes 是备注允许的字符数。
	maxRemarkRunes = 1024
	// maxMembers 是单条轮换允许的成员数。
	maxMembers = 50
	// scheduleDaysDefault 是未指定天数时的默认排班窗口。
	scheduleDaysDefault = 14
	// scheduleDaysMax 是单次排班窗口的上限。
	scheduleDaysMax = 30
	// scheduleShiftLimit 是单次返回班次数量的兜底上限。
	scheduleShiftLimit = 64
)

// Gateway 管理值班轮换，并按轮换配置计算当前与未来班次。
type Gateway struct {
	rotations repository.OnCallRotationRepo
	members   repository.MemberRepo
	access    access.Checker
	log       *slog.Logger
}

func NewGateway(rotations repository.OnCallRotationRepo, members repository.MemberRepo, checker access.Checker, logger *slog.Logger) *Gateway {
	return &Gateway{rotations: rotations, members: members, access: checker, log: logger}
}

// Create 校验并保存轮换；成员随轮换一起写入。
func (gateway *Gateway) Create(ctx context.Context, in bo.OnCallRotation) (*bo.OnCallRotation, error) {
	ns, err := gateway.access.Allow(ctx, perm.NotificationRotationWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := gateway.normalize(ctx, &in); err != nil {
		return nil, err
	}
	item, err := gateway.rotations.CreateRotation(ctx, in)
	return item, gateway.access.Wrap(err)
}

// Update 校验并整体替换成员顺序。
func (gateway *Gateway) Update(ctx context.Context, in bo.OnCallRotation) (*bo.OnCallRotation, error) {
	ns, err := gateway.access.Allow(ctx, perm.NotificationRotationWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := gateway.normalize(ctx, &in); err != nil {
		return nil, err
	}
	item, err := gateway.rotations.UpdateRotation(ctx, in)
	return item, gateway.access.Wrap(err)
}

// UpdateStatus 启用或停用轮换。停用后绑定它的 OnCall 只使用手填接收方。
func (gateway *Gateway) UpdateStatus(ctx context.Context, uid, status string) (*bo.OnCallRotation, error) {
	ns, err := gateway.access.Allow(ctx, perm.NotificationRotationWrite)
	if err != nil {
		return nil, err
	}
	if status != "ACTIVE" && status != "DISABLED" {
		return nil, merr.Invalid("invalid status")
	}
	item, err := gateway.rotations.UpdateRotationStatus(ctx, ns, uid, status)
	return item, gateway.access.Wrap(err)
}

// Delete 软删除轮换；已绑定它的 OnCall 在投递时按未绑定处理。
func (gateway *Gateway) Delete(ctx context.Context, uid string) error {
	ns, err := gateway.access.Allow(ctx, perm.NotificationRotationDelete)
	if err != nil {
		return err
	}
	return gateway.access.Wrap(gateway.rotations.DeleteRotation(ctx, ns, uid))
}

// Get 读取轮换详情。
func (gateway *Gateway) Get(ctx context.Context, uid string) (*bo.OnCallRotation, error) {
	ns, err := gateway.access.Allow(ctx, perm.NotificationRotationRead)
	if err != nil {
		return nil, err
	}
	item, err := gateway.rotations.GetRotation(ctx, ns, uid)
	return item, gateway.access.Wrap(err)
}

// List 分页列出轮换。
func (gateway *Gateway) List(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.OnCallRotation, int, int, int, error) {
	ns, err := gateway.access.Allow(ctx, perm.NotificationRotationRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := gateway.rotations.ListRotations(ctx, ns, strings.TrimSpace(name), p, s)
	if err != nil {
		return nil, 0, 0, 0, gateway.access.Wrap(err)
	}
	return items, total, p, s, nil
}

// Schedule 返回未来班次表：包含当前班次并标记当前值班，窗口默认 14 天。
func (gateway *Gateway) Schedule(ctx context.Context, uid string, days int32) ([]bo.OnCallRotationShift, error) {
	ns, err := gateway.access.Allow(ctx, perm.NotificationRotationRead)
	if err != nil {
		return nil, err
	}
	window := int(days)
	if window <= 0 {
		window = scheduleDaysDefault
	}
	if window > scheduleDaysMax {
		return nil, merr.Invalid("invalid schedule days")
	}
	item, err := gateway.rotations.GetRotation(ctx, ns, uid)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	shifts, err := shiftsOf(*item, time.Now().UTC(), window)
	if err != nil {
		return nil, err
	}
	return shifts, nil
}

// shiftsOf 用纯计算函数生成班次，并按 MemberUID 补上名称与联系方式。
// 时间窗按自然日推进；结束时间截到窗口末端。
func shiftsOf(item bo.OnCallRotation, now time.Time, days int) ([]bo.OnCallRotationShift, error) {
	rotation := oncall.Rotation{
		Timezone: item.Timezone, Period: item.Period, StartsAt: item.StartsAt,
		HandoffHour: item.HandoffHour,
	}
	rotation.Members = make([]oncall.RotationMember, 0, len(item.Members))
	byMember := make(map[string]bo.OnCallRotationMember, len(item.Members))
	for _, member := range item.Members {
		rotation.Members = append(rotation.Members, oncall.RotationMember{MemberUID: member.MemberUID})
		byMember[member.MemberUID] = member
	}
	shifts, err := rotation.Shifts(now, now.AddDate(0, 0, days), scheduleShiftLimit)
	if err != nil {
		return nil, merr.Invalid("invalid rotation configuration")
	}
	items := make([]bo.OnCallRotationShift, 0, len(shifts))
	for _, shift := range shifts {
		items = append(items, bo.OnCallRotationShift{
			StartAt: shift.StartAt, EndAt: shift.EndAt,
			Member: byMember[shift.Member.MemberUID], Current: shift.Current,
		})
	}
	return items, nil
}

// MyShifts 返回当前登录账号自己的班次：跨轮换汇总后按开始时间升序。
// 自服务接口只要求 oncall 读权限，不要求轮换管理权限——普通值班成员看不到自己的班，
// 这个视图就没有意义。不是当前租户的成员时返回空列表，而不是 403。
func (gateway *Gateway) MyShifts(ctx context.Context, days int32) ([]bo.MyOnCallShift, error) {
	ns, err := gateway.access.Allow(ctx, perm.NotificationOncallRead)
	if err != nil {
		return nil, err
	}
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	member, err := gateway.members.MemberByUser(ctx, ns, userUID)
	if err != nil {
		if merr.IsNotFound(err) {
			// 平台管理员等非租户成员没有个人班次，这不是错误。
			return []bo.MyOnCallShift{}, nil
		}
		return nil, gateway.access.Wrap(err)
	}
	window := int(days)
	if window <= 0 {
		window = scheduleDaysDefault
	}
	if window > scheduleDaysMax {
		return nil, merr.Invalid("invalid schedule days")
	}
	rotations, err := gateway.allRotations(ctx, ns)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	items := make([]bo.MyOnCallShift, 0, len(rotations))
	for _, rotation := range rotations {
		if rotation.Status != "ACTIVE" {
			continue
		}
		shifts, err := myShiftsOf(rotation, member.UID, now, window)
		if err != nil {
			return nil, err
		}
		items = append(items, shifts...)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].StartAt.Before(items[j].StartAt) })
	return items, nil
}

// allRotations 翻完整个轮换列表。个人视图必须服务端过滤：
// 只取前 100 条会把排在第 101 条的租户成员静默漏掉。
func (gateway *Gateway) allRotations(ctx context.Context, namespaceUID string) ([]bo.OnCallRotation, error) {
	const (
		pageSize = 100
		maxPages = 20
	)
	items := make([]bo.OnCallRotation, 0, pageSize)
	for current := 1; current <= maxPages; current++ {
		page, total, err := gateway.rotations.ListRotations(ctx, namespaceUID, "", current, pageSize)
		if err != nil {
			return nil, gateway.access.Wrap(err)
		}
		items = append(items, page...)
		if len(page) == 0 || len(items) >= total {
			break
		}
	}
	return items, nil
}

// myShiftsOf 从一次轮换的班次表里挑出属于我的班次。
func myShiftsOf(item bo.OnCallRotation, memberUID string, now time.Time, days int) ([]bo.MyOnCallShift, error) {
	shifts, err := shiftsOf(item, now, days)
	if err != nil {
		return nil, err
	}
	items := make([]bo.MyOnCallShift, 0, len(shifts))
	for _, shift := range shifts {
		if shift.Member.MemberUID != memberUID {
			continue
		}
		items = append(items, bo.MyOnCallShift{
			RotationUID:  item.UID,
			RotationName: item.Name,
			Timezone:     item.Timezone,
			StartAt:      shift.StartAt,
			EndAt:        shift.EndAt,
			Current:      shift.Current,
		})
	}
	return items, nil
}

// normalize 校验名称、时区、周期、交班时刻与成员，并规范化成员顺序。
// 成员必须是当前租户的 JOINED 成员，重复成员直接拒绝。
func (gateway *Gateway) normalize(ctx context.Context, in *bo.OnCallRotation) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || utf8.RuneCountInString(in.Name) > maxNameRunes {
		return merr.Invalid("invalid rotation name")
	}
	in.Remark = strings.TrimSpace(in.Remark)
	if utf8.RuneCountInString(in.Remark) > maxRemarkRunes {
		return merr.Invalid("remark is too long")
	}
	timezone, err := oncall.NormalizeTimezone(in.Timezone)
	if err != nil {
		return err
	}
	in.Timezone = timezone
	if in.Period != oncall.RotationPeriodDaily && in.Period != oncall.RotationPeriodWeekly {
		return merr.Invalid("invalid rotation period")
	}
	if in.HandoffHour < 0 || in.HandoffHour > 23 {
		return merr.Invalid("invalid rotation handoff hour")
	}
	if in.StartsAt.IsZero() {
		return merr.Invalid("rotation starts at is required")
	}
	if len(in.Members) == 0 || len(in.Members) > maxMembers {
		return merr.Invalid("invalid rotation members")
	}
	memberUIDs := make([]string, 0, len(in.Members))
	seen := make(map[string]struct{}, len(in.Members))
	members := make([]bo.OnCallRotationMember, 0, len(in.Members))
	for _, member := range in.Members {
		memberUID := strings.TrimSpace(member.MemberUID)
		if memberUID == "" {
			return merr.Invalid("rotation member is required")
		}
		if _, ok := seen[memberUID]; ok {
			return merr.Invalid("duplicate rotation member")
		}
		seen[memberUID] = struct{}{}
		memberUIDs = append(memberUIDs, memberUID)
		members = append(members, bo.OnCallRotationMember{MemberUID: memberUID})
	}
	rows, err := gateway.members.ListMembersByUIDs(ctx, in.NamespaceUID, memberUIDs)
	if err != nil {
		return gateway.access.Wrap(err)
	}
	joined := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row.Status == "JOINED" {
			joined[row.UID] = struct{}{}
		}
	}
	for index := range members {
		if _, ok := joined[members[index].MemberUID]; !ok {
			return merr.Invalid("invalid rotation member")
		}
		members[index].Position = index
	}
	in.Members = members
	return nil
}
