package user

import "testing"

func TestValidateUsername(t *testing.T) {
	valid := []string{"abc", "新用户", "user1234", "12345678", "名字abc"}
	for _, name := range valid {
		if err := ValidateUsername(name); err != nil {
			t.Fatalf("合法用户名 %q 被拒绝: %v", name, err)
		}
	}
	invalid := []string{"", "abcdefghi", "bad-name", "bad name", "名字!", "emoji😀"}
	for _, name := range invalid {
		if err := ValidateUsername(name); err == nil {
			t.Fatalf("非法用户名 %q 通过了校验", name)
		}
	}
}

func TestDefaultUsername(t *testing.T) {
	cases := map[string]string{
		"new@example.com":            "new",
		"john.doe123456@example.com": "johndoe1",
		"verylongprefix@example.com": "verylong",
		"___@example.com":            "user",
		"用户.名字@example.com":          "用户名字",
		"plain":                      "plain",
	}
	for email, want := range cases {
		got := DefaultUsername(email)
		if got != want {
			t.Fatalf("DefaultUsername(%q) = %q, want %q", email, got, want)
		}
		if err := ValidateUsername(got); err != nil {
			t.Fatalf("DefaultUsername(%q) = %q 不合规: %v", email, got, err)
		}
	}
}
