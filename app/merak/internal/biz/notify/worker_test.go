package notify

import (
	"testing"
	"time"
)

// TestReportSettledKeepsSuccessfulDelivery 确认投递成功不会被轮次截止或续租失败
// 翻成失败。控制面收到 OK=false 会把任务退回 PENDING 重新派发，等于重复通知。
func TestReportSettledKeepsSuccessfulDelivery(t *testing.T) {
	cases := []struct {
		name      string
		delivery  DeliveryResult
		interrupt string
		wantOK    bool
		want      string
	}{
		{name: "clean success", delivery: DeliveryResult{OK: true, Detail: "sent"}, wantOK: true, want: "sent"},
		{name: "success after round timeout", delivery: DeliveryResult{OK: true, Detail: "sent"}, interrupt: "round timed out", wantOK: true, want: "sent (round timed out)"},
		{name: "success after renew failure", delivery: DeliveryResult{OK: true, Detail: "sent"}, interrupt: "lock renew failed", wantOK: true, want: "sent (lock renew failed)"},
		{name: "success after stop", delivery: DeliveryResult{OK: true, Detail: "sent"}, interrupt: "worker stopped", wantOK: true, want: "sent (worker stopped)"},
		{name: "failed delivery keeps detail", delivery: DeliveryResult{Detail: "Post https://hook.test: context canceled"}, interrupt: "round timed out", wantOK: false, want: "Post https://hook.test: context canceled"},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			report := reportSettled("task-1", item.delivery, item.interrupt)
			if report.OK != item.wantOK {
				t.Fatalf("OK = %v, want %v", report.OK, item.wantOK)
			}
			if report.Detail != item.want {
				t.Fatalf("Detail = %q, want %q", report.Detail, item.want)
			}
			if report.TaskUID != "task-1" {
				t.Fatalf("TaskUID = %q", report.TaskUID)
			}
		})
	}
}

// TestRoundTimeoutOutlivesDeliveryTimeout 确认轮次截止严格晚于投递超时，
// 否则投递会被轮次提前打断，出现「请求已发出、本地按失败回写」的重复投递。
func TestRoundTimeoutOutlivesDeliveryTimeout(t *testing.T) {
	worker := NewWorker(nil, nil, nil, nil)
	if worker.timeout != 30*time.Second {
		t.Fatalf("default delivery timeout = %s, want 30s", worker.timeout)
	}
	if worker.roundTimeout != worker.timeout+roundOverhead {
		t.Fatalf("round timeout = %s, want %s", worker.roundTimeout, worker.timeout+roundOverhead)
	}
	if worker.roundTimeout <= worker.timeout {
		t.Fatalf("round timeout %s must exceed delivery timeout %s", worker.roundTimeout, worker.timeout)
	}
}
