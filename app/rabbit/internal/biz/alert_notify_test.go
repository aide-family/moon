package biz

import (
	"testing"

	"github.com/aide-family/rabbit/internal/biz/bo"
)

func TestAlertDedupFingerprint(t *testing.T) {
	payload := &bo.AlertPayloadBo{
		Fingerprint: "fp-1",
		Labels:      map[string]string{"alert_event_uid": "999"},
		GroupKey:    "gk",
	}
	if got := alertDedupFingerprint(payload); got != "fp-1" {
		t.Fatalf("fingerprint = %q, want fp-1", got)
	}
	payload = &bo.AlertPayloadBo{
		Labels: map[string]string{"alert_event_uid": "999"},
	}
	if got := alertDedupFingerprint(payload); got != "999" {
		t.Fatalf("fingerprint = %q, want alert_event_uid", got)
	}
}

func TestIsResolvedAlertStatus(t *testing.T) {
	if !isResolvedAlertStatus("resolved") {
		t.Fatal("expected resolved")
	}
	if isResolvedAlertStatus("firing") {
		t.Fatal("expected not resolved")
	}
}
