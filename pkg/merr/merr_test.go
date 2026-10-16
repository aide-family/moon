package merr

import (
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"
)

func TestCategories(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		code   int32
		reason string
	}{
		{name: "client", err: Client("bad"), code: 400, reason: ReasonClient},
		{name: "unauthenticated", err: Unauthenticated("login"), code: 401, reason: ReasonUnauthenticated},
		{name: "forbidden", err: Forbidden("denied"), code: 403, reason: ReasonForbidden},
		{name: "not found", err: NotFound("missing"), code: 404, reason: ReasonNotFound},
		{name: "conflict", err: Conflict("dup"), code: 409, reason: ReasonConflict},
		{name: "invalid", err: Invalid("bad"), code: 400, reason: ReasonInvalid},
		{name: "server", err: Server("boom"), code: 500, reason: ReasonServer},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := kerrors.FromError(tc.err)
			if got.Code != tc.code || got.Reason != tc.reason {
				t.Fatalf("code=%d reason=%s, want code=%d reason=%s", got.Code, got.Reason, tc.code, tc.reason)
			}
		})
	}
}

func TestClassification(t *testing.T) {
	if !IsNotFound(NotFound("missing")) || IsNotFound(Client("missing")) || IsNotFound(Conflict("missing")) {
		t.Fatal("IsNotFound")
	}
	if !IsConflict(Conflict("dup")) || IsConflict(NotFound("dup")) || IsConflict(Invalid("dup")) {
		t.Fatal("IsConflict")
	}
	if !IsInvalid(Invalid("bad")) || IsInvalid(Client("bad")) || IsInvalid(NotFound("bad")) {
		t.Fatal("IsInvalid")
	}
}
