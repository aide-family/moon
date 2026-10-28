package data

import (
	"context"
	"reflect"
	"testing"

	"github.com/aide-family/moon/app/alioth/internal/data/ent/member"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
)

// TestReplaceMemberUID 替换并去重；没命中时原样返回，避免无意义写库。
func TestReplaceMemberUID(t *testing.T) {
	cases := []struct {
		name    string
		values  []string
		want    []string
		changed bool
	}{
		{name: "replace and dedupe", values: []string{"m1", "m2", "m1"}, want: []string{"m9", "m2"}, changed: true},
		{name: "no match keeps original", values: []string{"a", "a"}, want: []string{"a", "a"}, changed: false},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			from, to := "m1", "m9"
			if item.name == "no match keeps original" {
				from = "missing"
			}
			got, changed := replaceMemberUID(item.values, from, to)
			if changed != item.changed {
				t.Fatalf("changed = %v, want %v", changed, item.changed)
			}
			if !reflect.DeepEqual(got, item.want) {
				t.Fatalf("got %v, want %v", got, item.want)
			}
		})
	}
}

// TestMergeAccountsRemapsNotificationGroupMembers 合并账号后，通知组引用的源成员 UID
// 必须改到目标成员，否则成员被软删后收件人会静默消失。
func TestMergeAccountsRemapsNotificationGroupMembers(t *testing.T) {
	fixture := newMergeFixture(t)
	source := fixture.createUser("grp-src", user.SystemRoleUSER)
	target := fixture.createUser("grp-dst", user.SystemRoleUSER)
	space, roles := fixture.createNamespace("grp", target.UID)
	srcMember := fixture.addMember(space, source.UID, member.StatusJOINED, roles[perm.RoleMember])
	dstMember := fixture.addMember(space, target.UID, member.StatusJOINED, roles[perm.RoleMember])

	group, err := fixture.store.client.NotificationGroup.Create().
		SetUID(id.New()).
		SetNamespaceUID(space.UID).
		SetName(fixture.prefix + "-group").
		SetEmailMemberUids([]string{srcMember.UID, dstMember.UID}).
		SetSmsMemberUids([]string{srcMember.UID}).
		SetMemberUids([]string{srcMember.UID}).
		Save(fixture.ctx())
	if err != nil {
		t.Fatalf("创建通知组: %v", err)
	}
	t.Cleanup(func() {
		_, _ = fixture.store.client.NotificationGroup.Delete().
			Where(notificationgroup.UIDEQ(group.UID)).
			Exec(context.Background())
	})

	if err := fixture.mergeAccounts(source.UID, target.UID); err != nil {
		t.Fatalf("MergeAccounts: %v", err)
	}

	got, err := fixture.store.client.NotificationGroup.Query().
		Where(notificationgroup.UIDEQ(group.UID)).
		Only(fixture.ctx())
	if err != nil {
		t.Fatalf("查询通知组: %v", err)
	}
	want := []string{dstMember.UID}
	if !reflect.DeepEqual(got.EmailMemberUids, want) {
		t.Fatalf("email_member_uids = %v, want %v", got.EmailMemberUids, want)
	}
	if !reflect.DeepEqual(got.SmsMemberUids, want) {
		t.Fatalf("sms_member_uids = %v, want %v", got.SmsMemberUids, want)
	}
	if !reflect.DeepEqual(got.MemberUids, want) {
		t.Fatalf("member_uids = %v, want %v", got.MemberUids, want)
	}

	// 源成员被软删，目标成员保留。
	if fixture.activeMember(space.UID, source.UID) != nil {
		t.Fatal("source member should be soft-deleted after merge")
	}
	if fixture.activeMember(space.UID, target.UID) == nil {
		t.Fatal("target member should remain active after merge")
	}
}
