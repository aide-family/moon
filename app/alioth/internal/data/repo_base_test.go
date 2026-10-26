package data

import (
	"errors"
	"testing"
)

// TestConstraintMessage 唯一约束错误按约束名给出可读文案，未知约束保持通用提示。
func TestConstraintMessage(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "oncall group binding",
			err:  errors.New(`pq: duplicate key value violates unique constraint "notificationoncallgroup_oncall_uid_group_uid"`),
			want: "notification group is already bound to this on-call",
		},
		{
			name: "user email",
			err:  errors.New(`pq: duplicate key value violates unique constraint "users_email_key"`),
			want: "email already exists",
		},
		{
			name: "member",
			err:  errors.New(`pq: duplicate key value violates unique constraint "member_namespace_uid_user_uid"`),
			want: "member already exists",
		},
		{
			name: "unknown",
			err:  errors.New(`pq: duplicate key value violates unique constraint "some_other_key"`),
			want: "name already exists",
		},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			if got := constraintMessage(item.err); got != item.want {
				t.Fatalf("constraintMessage = %q, want %q", got, item.want)
			}
		})
	}
}
