package server

import (
	"context"
	"strconv"
	"testing"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
)

func TestProtoValidatorRejectsInvalidRequest(t *testing.T) {
	handler := func(context.Context, any) (any, error) { return nil, nil }
	middleware := protoValidator()
	invalid := &v1.RegisterRequest{Email: "not-an-email", Name: "user", Code: "1234", CaptchaId: "id", CaptchaCode: "12"}
	if _, err := middleware(handler)(context.Background(), invalid); err == nil {
		t.Fatal("invalid register request passed validation")
	}
	valid := &v1.RegisterRequest{Email: "user@example.com", Name: "新用户1", Code: "1234", CaptchaId: "id", CaptchaCode: "12"}
	if _, err := middleware(handler)(context.Background(), valid); err != nil {
		t.Fatalf("valid register request rejected: %v", err)
	}
	shortCode := &v1.RegisterRequest{Email: "user@example.com", Name: "user", Code: "123", CaptchaId: "id", CaptchaCode: "12"}
	if _, err := middleware(handler)(context.Background(), shortCode); err == nil {
		t.Fatal("short email code passed validation")
	}
	missingCaptcha := &v1.RegisterRequest{Email: "user@example.com", Name: "user", Code: "1234"}
	if _, err := middleware(handler)(context.Background(), missingCaptcha); err == nil {
		t.Fatal("register request without captcha passed validation")
	}
}

// TestProtoValidatorChecksUsername 确认用户名规则同时落在接口契约上：
// 只允许中英文和数字，最多 8 个字符；留空由后端从邮箱前缀派生。
func TestProtoValidatorChecksUsername(t *testing.T) {
	handler := func(context.Context, any) (any, error) { return nil, nil }
	middleware := protoValidator()
	for _, name := range []string{"abcdefghi", "bad-name", "bad name", "名字!"} {
		request := &v1.RegisterRequest{Email: "user@example.com", Name: name, Code: "1234", CaptchaId: "id", CaptchaCode: "12"}
		if _, err := middleware(handler)(context.Background(), request); err == nil {
			t.Fatalf("用户名 %q 通过了接口校验", name)
		}
	}
	for _, name := range []string{"user1234", "新用户", ""} {
		request := &v1.RegisterRequest{Email: "user@example.com", Name: name, Code: "1234", CaptchaId: "id", CaptchaCode: "12"}
		if _, err := middleware(handler)(context.Background(), request); err != nil {
			t.Fatalf("合法用户名 %q 被拒绝: %v", name, err)
		}
	}
}

func TestProtoValidatorRejectsEmptyRoleList(t *testing.T) {
	handler := func(context.Context, any) (any, error) { return nil, nil }
	request := &v1.UpdateMemberRolesRequest{Uid: "member-1", RoleUids: nil}
	if _, err := protoValidator()(handler)(context.Background(), request); err == nil {
		t.Fatal("empty role list passed validation")
	}
}

func TestProtoValidatorIgnoresNonProtoRequests(t *testing.T) {
	handler := func(context.Context, any) (any, error) { return nil, nil }
	if _, err := protoValidator()(handler)(context.Background(), "plain"); err != nil {
		t.Fatalf("non-proto request rejected: %v", err)
	}
}

func TestProtoValidatorLimitsRealtimeTabs(t *testing.T) {
	handler := func(context.Context, any) (any, error) { return nil, nil }
	uids := make([]string, 11)
	for index := range uids {
		uids[index] = "page-" + strconv.Itoa(index)
	}
	if _, err := protoValidator()(handler)(context.Background(), &v1.SetMyRealtimeTabsRequest{AlertPageUids: uids}); err == nil {
		t.Fatal("11 alert pages passed validation")
	}
	if _, err := protoValidator()(handler)(context.Background(), &v1.SetMyRealtimeTabsRequest{AlertPageUids: uids[:10]}); err != nil {
		t.Fatalf("10 alert pages rejected: %v", err)
	}
}

func TestProtoValidatorAllowsPermanentSuppress(t *testing.T) {
	handler := func(context.Context, any) (any, error) { return nil, nil }
	if _, err := protoValidator()(handler)(context.Background(), &v1.SuppressAlertEventRequest{Uid: "event-1", Permanent: true}); err != nil {
		t.Fatalf("permanent suppress rejected: %v", err)
	}
	if _, err := protoValidator()(handler)(context.Background(), &v1.SuppressAlertEventRequest{Uid: "event-1", SuppressedSeconds: 60}); err != nil {
		t.Fatalf("timed suppress rejected: %v", err)
	}
}
