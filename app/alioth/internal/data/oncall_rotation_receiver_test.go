package data

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/member"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationchannel"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationoncall"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationoncallgroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationtemplate"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/oncallrotation"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/oncallrotationmember"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
)

// newRotationTestStore 连接集成测试库并应用 schema；没有 ALIOTH_TEST_DSN 时跳过。
func newRotationTestStore(t *testing.T) (*Dependencies, context.Context) {
	t.Helper()
	source := testDSN(t)
	ctx := context.Background()
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Skip(err)
	}
	t.Cleanup(cleanup)
	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}
	return store, ctx
}

// cleanupRotationNamespace 清掉用例写入的轮换、OnCall 与通知配置。
func cleanupRotationNamespace(store *Dependencies, namespace string, userUIDs []string) func() {
	return func() {
		ctx := context.Background()
		_, _ = store.client.NotificationTask.Delete().Where(notificationtask.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.NotificationOnCallGroup.Delete().Where(notificationoncallgroup.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.NotificationOnCall.Delete().Where(notificationoncall.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.NotificationGroup.Delete().Where(notificationgroup.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.NotificationChannel.Delete().Where(notificationchannel.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.NotificationTemplate.Delete().Where(notificationtemplate.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.OnCallRotationMember.Delete().Where(oncallrotationmember.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.OnCallRotation.Delete().Where(oncallrotation.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.Member.Delete().Where(member.NamespaceUIDEQ(namespace)).Exec(ctx)
		if len(userUIDs) > 0 {
			_, _ = store.client.User.Delete().Where(user.UIDIn(userUIDs...)).Exec(ctx)
		}
	}
}

// rotationTestMember 创建一个带邮箱的平台用户与 JOINED 租户成员。
func rotationTestMember(ctx context.Context, t *testing.T, store *Dependencies, namespace, email string) (string, string) {
	t.Helper()
	account, err := store.client.User.Create().
		SetUID(id.New()).
		SetName(strings.Split(email, "@")[0]).
		SetEmail(email).
		SetStatus(user.StatusACTIVE).
		Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	row, err := store.client.Member.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespace).
		SetUserUID(account.UID).
		SetStatus(member.StatusJOINED).
		Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return row.UID, account.UID
}

// rotationTestFixture 创建渠道、模板、通知组与绑定该组的 OnCall。
func rotationTestFixture(ctx context.Context, t *testing.T, store *Dependencies, namespace, groupName, manualEmail, rotationUID string) string {
	t.Helper()
	channel, err := store.client.NotificationChannel.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespace).
		SetName("轮换邮件渠道-" + groupName).
		SetType(notificationchannel.TypeSMTP).
		SetStatus(notificationchannel.StatusACTIVE).
		SetFromAddress("ops@example.com").
		Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	template, err := store.client.NotificationTemplate.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespace).
		SetName("轮换邮件模板-" + groupName).
		SetType(notificationtemplate.TypeEMAIL).
		SetStatus(notificationtemplate.StatusACTIVE).
		SetSubject("[{{.phase}}] {{.summary}}").
		SetBody("{{.summary}}").
		Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	group, err := store.client.NotificationGroup.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespace).
		SetName(groupName).
		SetStatus(notificationgroup.StatusACTIVE).
		SetChannelUids([]string{channel.UID}).
		SetTemplateUids([]string{template.UID}).
		SetEmailReceivers(manualEmail).
		Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewNotificationRepo(store, slog.Default())
	if _, err := repo.CreateOnCall(ctx, bo.NotificationOnCall{
		NamespaceUID: namespace, Name: "oncall-" + groupName, Labels: "{}", Timezone: "Asia/Shanghai",
		Groups: []bo.NotificationOnCallGroup{{GroupUID: group.UID, RotationUID: rotationUID}},
	}); err != nil {
		t.Fatal(err)
	}
	return group.UID
}

// dispatchTestEvent 用同一个事务里的分发链路生成 FIRING 通知任务，返回事件 UID。
func dispatchTestEvent(ctx context.Context, t *testing.T, store *Dependencies, namespace, eventUID string) {
	t.Helper()
	tx, err := store.client.Tx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := dispatchEventNotifications(ctx, tx, notifyEvent{
		NamespaceUID: namespace, EventUID: eventUID, RuleUID: "rule-rotation", DatasourceUID: "ds-rotation",
		Phase: "FIRING", Summary: "轮换接收方测试", Labels: "{}", Fingerprint: "fp-" + eventUID,
	}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

// taskReceivers 返回某事件某个通知组生成的任务接收方。
func taskReceivers(ctx context.Context, t *testing.T, store *Dependencies, eventUID, groupUID string) string {
	t.Helper()
	row, err := store.client.NotificationTask.Query().Where(
		notificationtask.EventUIDEQ(eventUID),
		notificationtask.GroupUIDEQ(groupUID),
	).First(ctx)
	if err != nil {
		t.Fatalf("task for group %s: %v", groupUID, err)
	}
	return row.Receivers
}

// TestRotationReceiverMerge 覆盖轮换绑定到 OnCall 后的接收方合并：
// 当班成员邮箱与手填邮箱合并去重；轮换停用时只用固定接收方；重复邮箱只保留一次。
func TestRotationReceiverMerge(t *testing.T) {
	store, ctx := newRotationTestStore(t)
	namespace := "ns-rotation-" + time.Now().Format("150405.000000")
	otherNamespace := namespace + "-other"
	createdUsers := make([]string, 0, 2)
	t.Cleanup(cleanupRotationNamespace(store, namespace, createdUsers))
	t.Cleanup(func() {
		ctx := context.Background()
		_, _ = store.client.OnCallRotationMember.Delete().Where(oncallrotationmember.NamespaceUIDEQ(otherNamespace)).Exec(ctx)
		_, _ = store.client.OnCallRotation.Delete().Where(oncallrotation.NamespaceUIDEQ(otherNamespace)).Exec(ctx)
		_, _ = store.client.Member.Delete().Where(member.NamespaceUIDEQ(otherNamespace)).Exec(ctx)
	})
	repo := NewOnCallRotationRepo(store, slog.Default())

	manualEmailA := "manual-a-" + namespace + "@example.com"
	manualEmailB := "manual-b-" + namespace + "@example.com"
	manualEmailC := "manual-c-" + namespace + "@example.com"
	rotatorEmail := "rotator-" + namespace + "@example.com"
	rotatorMember, rotatorUser := rotationTestMember(ctx, t, store, namespace, rotatorEmail)
	manualMember, manualUser := rotationTestMember(ctx, t, store, namespace, manualEmailB)
	createdUsers = append(createdUsers, rotatorUser, manualUser)

	newRotation := func(name, memberUID string) *bo.OnCallRotation {
		t.Helper()
		rotation, err := repo.CreateRotation(ctx, bo.OnCallRotation{
			NamespaceUID: namespace, Name: name, Timezone: "Asia/Shanghai", Period: "DAILY",
			StartsAt: time.Now().UTC().Add(-24 * time.Hour), HandoffHour: 0,
			Members: []bo.OnCallRotationMember{{MemberUID: memberUID}},
		})
		if err != nil {
			t.Fatal(err)
		}
		return rotation
	}

	rotatorRotation := newRotation("主轮换", rotatorMember)
	manualRotation := newRotation("重复邮箱轮换", manualMember)
	disabledRotation := newRotation("停用轮换", rotatorMember)
	if _, err := repo.UpdateRotationStatus(ctx, namespace, disabledRotation.UID, "DISABLED"); err != nil {
		t.Fatal(err)
	}

	rotatorGroup := rotationTestFixture(ctx, t, store, namespace, "轮换组", manualEmailA, rotatorRotation.UID)
	manualGroup := rotationTestFixture(ctx, t, store, namespace, "重复邮箱组", manualEmailB, manualRotation.UID)
	disabledGroup := rotationTestFixture(ctx, t, store, namespace, "停用轮换组", manualEmailC, disabledRotation.UID)

	eventUID := id.New()
	// 绑定非本租户的轮换必须拒绝。
	foreignMember, foreignUser := rotationTestMember(ctx, t, store, otherNamespace, "foreign-"+namespace+"@example.com")
	createdUsers = append(createdUsers, foreignUser)
	foreign := newRotationForOtherNamespace(t, repo, otherNamespace, foreignMember)
	if _, err := NewNotificationRepo(store, slog.Default()).CreateOnCall(ctx, bo.NotificationOnCall{
		NamespaceUID: namespace, Name: "非法绑定", Labels: "{}", Timezone: "Asia/Shanghai",
		Groups: []bo.NotificationOnCallGroup{{GroupUID: rotatorGroup, RotationUID: foreign}},
	}); err == nil {
		t.Fatal("binding a rotation from another namespace should fail")
	}

	dispatchTestEvent(ctx, t, store, namespace, eventUID)

	if got := taskReceivers(ctx, t, store, eventUID, rotatorGroup); got != manualEmailA+","+rotatorEmail {
		t.Fatalf("rotator group receivers = %q, want %q", got, manualEmailA+","+rotatorEmail)
	}
	if got := taskReceivers(ctx, t, store, eventUID, manualGroup); got != manualEmailB {
		t.Fatalf("duplicate email group receivers = %q, want deduplicated fixed receiver", got)
	}
	if got := taskReceivers(ctx, t, store, eventUID, disabledGroup); got != manualEmailC {
		t.Fatalf("disabled rotation group receivers = %q, want fixed receiver only", got)
	}
}

// newRotationForOtherNamespace 在另一个租户下建一条轮换，用于跨租户绑定校验。
func newRotationForOtherNamespace(t *testing.T, repo repository.OnCallRotationRepo, namespace, memberUID string) string {
	t.Helper()
	rotation, err := repo.CreateRotation(context.Background(), bo.OnCallRotation{
		NamespaceUID: namespace, Name: "其它租户轮换", Timezone: "Asia/Shanghai", Period: "DAILY",
		StartsAt: time.Now().UTC(), HandoffHour: 0,
		Members: []bo.OnCallRotationMember{{MemberUID: memberUID}},
	})
	if err != nil {
		// 成员校验在 biz；仓储允许任意 member_uid，因此这里应能创建成功。
		t.Fatal(err)
	}
	return rotation.UID
}
