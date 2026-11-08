package server

import (
	"testing"
	"time"
)

func TestPublicReportLimiter(t *testing.T) {
	limiter := newPublicReportLimiter()
	now := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	for i := 0; i < int(publicReportBurst); i++ {
		if !limiter.allow("token", now) {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	if limiter.allow("token", now) {
		t.Fatal("request over burst should be limited")
	}
	if !limiter.allow("other", now) {
		t.Fatal("different token must have its own bucket")
	}
	// 30 次/秒，等待 2 秒可恢复 60 个令牌，但上限是 burst。
	if !limiter.allow("token", now.Add(2*time.Second)) {
		t.Fatal("bucket should refill over time")
	}
}
