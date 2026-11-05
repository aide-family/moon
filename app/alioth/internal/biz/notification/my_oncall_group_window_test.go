package notification

import (
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

func myOnCallTestLocation(t *testing.T) *time.Location {
	t.Helper()
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load Asia/Shanghai: %v", err)
	}
	return location
}

func myOnCallActive(uid, name, timezone string, groups ...bo.NotificationOnCallGroup) bo.NotificationOnCall {
	return bo.NotificationOnCall{
		UID: uid, Name: name, Status: "ACTIVE", Timezone: timezone, Groups: groups,
	}
}

func myOnCallGroup(uid, name, status string) bo.NotificationGroup {
	return bo.NotificationGroup{UID: uid, Name: name, Status: status}
}

func TestMyOnCallGroupWindows(t *testing.T) {
	location := myOnCallTestLocation(t)
	mondayNight := time.Date(2026, time.November, 2, 23, 30, 0, 0, location)
	if mondayNight.Weekday() != time.Monday {
		t.Fatalf("fixture day = %s, want Monday", mondayNight.Weekday())
	}
	nightWindow := []bo.NotificationOnCallWindow{
		{Weekdays: []string{"MONDAY"}, Start: "22:00", End: "08:00"},
	}

	// 仓储按 groupMatchesUser 过滤后的结果在这里只按 UID 取交集；
	// 成员 UID / 邮箱 / 手机号三种命中路径由 data 包的 groupMatchesUser 单测覆盖。
	t.Run("成员 UID 命中的通知组返回条目", func(t *testing.T) {
		onCall := myOnCallActive("oncall-1", "夜间值班", "Asia/Shanghai",
			bo.NotificationOnCallGroup{GroupUID: "group-1", Windows: nightWindow})
		items, err := myOnCallGroupWindows([]bo.NotificationOnCall{onCall}, []bo.NotificationGroup{myOnCallGroup("group-1", "值班组", "ACTIVE")}, mondayNight)
		if err != nil {
			t.Fatalf("windows: %v", err)
		}
		if len(items) != 1 || items[0].GroupUID != "group-1" || items[0].GroupName != "值班组" {
			t.Fatalf("items = %+v, want group-1", items)
		}
	})

	t.Run("仅邮箱命中的通知组返回条目", func(t *testing.T) {
		onCall := myOnCallActive("oncall-1", "夜间值班", "Asia/Shanghai",
			bo.NotificationOnCallGroup{GroupUID: "group-email"})
		items, err := myOnCallGroupWindows([]bo.NotificationOnCall{onCall}, []bo.NotificationGroup{myOnCallGroup("group-email", "邮箱值班组", "ACTIVE")}, mondayNight)
		if err != nil {
			t.Fatalf("windows: %v", err)
		}
		if len(items) != 1 || items[0].GroupUID != "group-email" {
			t.Fatalf("items = %+v, want group-email", items)
		}
	})

	t.Run("仅手机号命中的通知组返回条目", func(t *testing.T) {
		onCall := myOnCallActive("oncall-1", "夜间值班", "Asia/Shanghai",
			bo.NotificationOnCallGroup{GroupUID: "group-sms"})
		items, err := myOnCallGroupWindows([]bo.NotificationOnCall{onCall}, []bo.NotificationGroup{myOnCallGroup("group-sms", "短信值班组", "ACTIVE")}, mondayNight)
		if err != nil {
			t.Fatalf("windows: %v", err)
		}
		if len(items) != 1 || items[0].GroupUID != "group-sms" {
			t.Fatalf("items = %+v, want group-sms", items)
		}
	})

	t.Run("未绑定到我的通知组不返回", func(t *testing.T) {
		onCall := myOnCallActive("oncall-1", "夜间值班", "Asia/Shanghai",
			bo.NotificationOnCallGroup{GroupUID: "group-other"})
		items, err := myOnCallGroupWindows([]bo.NotificationOnCall{onCall}, []bo.NotificationGroup{myOnCallGroup("group-1", "值班组", "ACTIVE")}, mondayNight)
		if err != nil {
			t.Fatalf("windows: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("items = %+v, want empty", items)
		}
	})

	t.Run("跨午夜窗口按开始日判断", func(t *testing.T) {
		onCall := myOnCallActive("oncall-1", "夜间值班", "Asia/Shanghai",
			bo.NotificationOnCallGroup{GroupUID: "group-1", Windows: nightWindow})
		groups := []bo.NotificationGroup{myOnCallGroup("group-1", "值班组", "ACTIVE")}
		cases := []struct {
			name string
			at   time.Time
			want bool
		}{
			{name: "开始日深夜", at: mondayNight, want: true},
			{name: "次日凌晨", at: time.Date(2026, time.November, 3, 1, 0, 0, 0, location), want: true},
			{name: "开始日白天", at: time.Date(2026, time.November, 2, 12, 0, 0, 0, location), want: false},
			{name: "结束后", at: time.Date(2026, time.November, 3, 8, 0, 0, 0, location), want: false},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				items, err := myOnCallGroupWindows([]bo.NotificationOnCall{onCall}, groups, tc.at)
				if err != nil {
					t.Fatalf("windows: %v", err)
				}
				if len(items) != 1 || items[0].Current != tc.want {
					t.Fatalf("current = %+v, want %v", items, tc.want)
				}
			})
		}
	})

	t.Run("空窗口表示全天候", func(t *testing.T) {
		onCall := myOnCallActive("oncall-1", "全天值班", "Asia/Shanghai",
			bo.NotificationOnCallGroup{GroupUID: "group-1"})
		items, err := myOnCallGroupWindows([]bo.NotificationOnCall{onCall}, []bo.NotificationGroup{myOnCallGroup("group-1", "值班组", "ACTIVE")}, mondayNight)
		if err != nil {
			t.Fatalf("windows: %v", err)
		}
		if len(items) != 1 || !items[0].Current {
			t.Fatalf("items = %+v, want all-day current", items)
		}
	})

	t.Run("停用的值班项被排除", func(t *testing.T) {
		onCall := myOnCallActive("oncall-1", "夜间值班", "Asia/Shanghai",
			bo.NotificationOnCallGroup{GroupUID: "group-1", Windows: nightWindow})
		onCall.Status = "DISABLED"
		items, err := myOnCallGroupWindows([]bo.NotificationOnCall{onCall}, []bo.NotificationGroup{myOnCallGroup("group-1", "值班组", "ACTIVE")}, mondayNight)
		if err != nil {
			t.Fatalf("windows: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("items = %+v, want empty", items)
		}
	})

	t.Run("停用的通知组被排除", func(t *testing.T) {
		onCall := myOnCallActive("oncall-1", "夜间值班", "Asia/Shanghai",
			bo.NotificationOnCallGroup{GroupUID: "group-1", Windows: nightWindow})
		items, err := myOnCallGroupWindows([]bo.NotificationOnCall{onCall}, []bo.NotificationGroup{myOnCallGroup("group-1", "值班组", "DISABLED")}, mondayNight)
		if err != nil {
			t.Fatalf("windows: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("items = %+v, want empty", items)
		}
	})

	t.Run("非成员没有匹配组时返回空结果", func(t *testing.T) {
		onCall := myOnCallActive("oncall-1", "夜间值班", "Asia/Shanghai",
			bo.NotificationOnCallGroup{GroupUID: "group-1", Windows: nightWindow})
		items, err := myOnCallGroupWindows([]bo.NotificationOnCall{onCall}, nil, mondayNight)
		if err != nil {
			t.Fatalf("windows: %v", err)
		}
		if items == nil || len(items) != 0 {
			t.Fatalf("items = %+v, want non-nil empty", items)
		}
	})

	t.Run("按值班项名称和通知组名称排序", func(t *testing.T) {
		onCalls := []bo.NotificationOnCall{
			myOnCallActive("oncall-b", "B 值班", "Asia/Shanghai",
				bo.NotificationOnCallGroup{GroupUID: "group-b2"},
				bo.NotificationOnCallGroup{GroupUID: "group-b1"}),
			myOnCallActive("oncall-a", "A 值班", "Asia/Shanghai",
				bo.NotificationOnCallGroup{GroupUID: "group-a"}),
		}
		groups := []bo.NotificationGroup{
			myOnCallGroup("group-b2", "beta", "ACTIVE"),
			myOnCallGroup("group-b1", "alpha", "ACTIVE"),
			myOnCallGroup("group-a", "alpha", "ACTIVE"),
		}
		items, err := myOnCallGroupWindows(onCalls, groups, mondayNight)
		if err != nil {
			t.Fatalf("windows: %v", err)
		}
		want := []string{"A 值班/alpha", "B 值班/alpha", "B 值班/beta"}
		if len(items) != len(want) {
			t.Fatalf("items = %+v, want %d", items, len(want))
		}
		for i, item := range items {
			got := item.OnCallName + "/" + item.GroupName
			if got != want[i] {
				t.Fatalf("item %d = %q, want %q", i, got, want[i])
			}
		}
	})

	t.Run("损坏的时段按服务端错误返回", func(t *testing.T) {
		onCall := myOnCallActive("oncall-1", "夜间值班", "Asia/Shanghai",
			bo.NotificationOnCallGroup{
				GroupUID: "group-1",
				Windows:  []bo.NotificationOnCallWindow{{Weekdays: []string{"FUNDAY"}, Start: "22:00", End: "08:00"}},
			})
		_, err := myOnCallGroupWindows([]bo.NotificationOnCall{onCall}, []bo.NotificationGroup{myOnCallGroup("group-1", "值班组", "ACTIVE")}, mondayNight)
		if kerrors.Reason(err) != merr.ReasonServer {
			t.Fatalf("reason = %q, want %q (err = %v)", kerrors.Reason(err), merr.ReasonServer, err)
		}
	})
}

func TestMemberContactTokens(t *testing.T) {
	tokens := memberContactTokens(&bo.Member{UID: "member-1", Email: "alice@example.com", Phone: "13800000000"})
	if len(tokens) != 2 || tokens[0] != "alice@example.com" || tokens[1] != "13800000000" {
		t.Fatalf("tokens = %v, want email and phone", tokens)
	}
	if got := memberContactTokens(&bo.Member{UID: "member-1"}); len(got) != 0 {
		t.Fatalf("tokens = %v, want empty", got)
	}
}
