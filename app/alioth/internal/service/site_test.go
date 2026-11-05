package service

import (
	"testing"
	"time"
)

// 乐观锁令牌必须能原样往返：带微秒的时间截到秒后再比较会永远不相等，
// 表现就是「GET 详情后立刻保存」也报 site page was modified。
func TestParseOptionalTimeRoundTrip(t *testing.T) {
	original := time.Date(2026, 11, 5, 12, 0, 0, 123456000, time.UTC)
	raw := formatLockTime(original)
	parsed, err := parseOptionalTime(raw)
	if err != nil {
		t.Fatalf("parse %q: %v", raw, err)
	}
	if parsed == nil || !parsed.Equal(original) {
		t.Fatalf("parsed = %v, want %v", parsed, original)
	}
}

// 旧客户端只传秒级时间也要能解析，不能因此拒绝保存。
func TestParseOptionalTimeAcceptsSeconds(t *testing.T) {
	parsed, err := parseOptionalTime("2026-11-05T12:00:00Z")
	if err != nil || parsed == nil {
		t.Fatalf("parse seconds: %v, %v", parsed, err)
	}
}

func TestParseOptionalTimeRejectsGarbage(t *testing.T) {
	if _, err := parseOptionalTime("not-a-time"); err == nil {
		t.Fatal("garbage timestamp should be rejected")
	}
	if parsed, err := parseOptionalTime("  "); err != nil || parsed != nil {
		t.Fatalf("blank timestamp should mean no lock: %v, %v", parsed, err)
	}
}
