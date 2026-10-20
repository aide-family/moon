package server

import (
	"context"
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
