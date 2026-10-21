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
	invalid := &v1.RegisterRequest{Email: "not-an-email", Password: "short"}
	if _, err := middleware(handler)(context.Background(), invalid); err == nil {
		t.Fatal("invalid register request passed validation")
	}
	valid := &v1.RegisterRequest{Email: "user@example.com", Password: "password1", CaptchaId: "id", CaptchaCode: "ABCD"}
	if _, err := middleware(handler)(context.Background(), valid); err != nil {
		t.Fatalf("valid register request rejected: %v", err)
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
