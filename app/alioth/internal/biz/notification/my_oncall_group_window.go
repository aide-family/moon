package notification

import (
	"context"
	"sort"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/oncall"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	// myOnCallPageSize 与值班路由列表的每页上限一致。
	myOnCallPageSize = 100
	// myOnCallMaxPages 是个人视图遍历值班路由的兜底页数，避免异常分页导致死循环。
	myOnCallMaxPages = 20
)

// MyOnCallGroupWindows 返回当前登录账号所在通知组在值班路由上的接收时段。
//
// 个人视图必须服务端过滤：只返回 ACTIVE 且未删除的值班路由与通知组；
// 当前账号不是租户成员（如平台管理员）时没有个人接收时段，返回空列表而不是拒绝。
func (center *Center) MyOnCallGroupWindows(ctx context.Context) ([]bo.MyOnCallGroupWindow, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationOncallRead)
	if err != nil {
		return nil, err
	}
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	member, err := center.members.MemberByUser(ctx, ns, userUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return []bo.MyOnCallGroupWindow{}, nil
		}
		return nil, center.access.Wrap(err)
	}
	// 成员 UID 由仓储单独匹配；tokens 只放手填接收方可能使用的邮箱和手机号。
	groups, err := center.repo.ListGroupsForMember(ctx, ns, member.UID, memberContactTokens(member))
	if err != nil {
		return nil, center.access.Wrap(err)
	}
	onCalls, err := center.allOnCalls(ctx, ns)
	if err != nil {
		return nil, err
	}
	return myOnCallGroupWindows(onCalls, groups, time.Now())
}

// allOnCalls 翻完整个值班路由列表。个人视图需要全量过滤，
// 只取第一页会把名称排在后面的路由静默漏掉。
func (center *Center) allOnCalls(ctx context.Context, namespaceUID string) ([]bo.NotificationOnCall, error) {
	items := make([]bo.NotificationOnCall, 0, myOnCallPageSize)
	for current := 1; current <= myOnCallMaxPages; current++ {
		pageItems, total, err := center.repo.ListOnCalls(ctx, namespaceUID, current, myOnCallPageSize, "")
		if err != nil {
			return nil, center.access.Wrap(err)
		}
		items = append(items, pageItems...)
		if len(pageItems) == 0 || len(items) >= total {
			break
		}
	}
	return items, nil
}

// memberContactTokens 返回用于匹配手填接收方的账号标识。
func memberContactTokens(member *bo.Member) []string {
	tokens := make([]string, 0, 2)
	if member.Email != "" {
		tokens = append(tokens, member.Email)
	}
	if member.Phone != "" {
		tokens = append(tokens, member.Phone)
	}
	return tokens
}

// myOnCallGroupWindows 把值班路由和已匹配到当前成员的通知组合并成个人视图条目。
// 纯函数：不访问仓储和权限，便于单测；通知组是否属于当前成员由调用方用
// data 包的 groupMatchesUser 判定，这里只按 UID 取交集。
func myOnCallGroupWindows(onCalls []bo.NotificationOnCall, groups []bo.NotificationGroup, at time.Time) ([]bo.MyOnCallGroupWindow, error) {
	myGroups := make(map[string]bo.NotificationGroup, len(groups))
	for _, group := range groups {
		// 仓储已只返回 ACTIVE，这里再挡一层，防止调用方调整查询条件后把停用组带出来。
		if group.UID == "" || group.Status != "ACTIVE" {
			continue
		}
		myGroups[group.UID] = group
	}
	items := make([]bo.MyOnCallGroupWindow, 0, len(onCalls))
	for _, onCall := range onCalls {
		if onCall.Status != "ACTIVE" {
			continue
		}
		for _, binding := range onCall.Groups {
			group, ok := myGroups[binding.GroupUID]
			if !ok {
				continue
			}
			current, err := onCallGroupCurrent(onCall, binding.Windows, at)
			if err != nil {
				return nil, err
			}
			items = append(items, bo.MyOnCallGroupWindow{
				OnCallUID:  onCall.UID,
				OnCallName: onCall.Name,
				GroupUID:   group.UID,
				GroupName:  group.Name,
				Timezone:   onCall.Timezone,
				Current:    current,
				Windows:    append([]bo.NotificationOnCallWindow(nil), binding.Windows...),
			})
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].OnCallName != items[j].OnCallName {
			return items[i].OnCallName < items[j].OnCallName
		}
		return items[i].GroupName < items[j].GroupName
	})
	return items, nil
}

// onCallGroupCurrent 按值班路由的时区判断 at 是否落在该通知组的任一接收时段内。
// 空窗口表示全天候；Compile 失败说明保存的时段已损坏，按服务端错误返回。
func onCallGroupCurrent(onCall bo.NotificationOnCall, windows []bo.NotificationOnCallWindow, at time.Time) (bool, error) {
	compiled, err := oncall.Compile(onCallSpecs(windows))
	if err != nil {
		return false, merr.Server("invalid on-call time window")
	}
	location, err := oncall.LoadLocation(onCall.Timezone)
	if err != nil {
		return false, merr.Server("invalid on-call timezone")
	}
	return oncall.Active(compiled, at.In(location)), nil
}
