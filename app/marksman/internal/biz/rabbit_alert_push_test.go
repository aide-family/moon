package biz

import (
	"context"
	"errors"
	"testing"

	rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
)

type stubAlertPushDedup struct {
	exists bool
	marked bool
	cleared bool
}

func (s *stubAlertPushDedup) ExistsFiringPush(_ context.Context, _ snowflake.ID, _ string) (bool, error) {
	return s.exists, nil
}

func (s *stubAlertPushDedup) MarkFiringPush(_ context.Context, _ snowflake.ID, _ string) error {
	s.marked = true
	return nil
}

func (s *stubAlertPushDedup) ClearFiringPush(_ context.Context, _ snowflake.ID, _ string) error {
	s.cleared = true
	return nil
}

type stubRabbitAlertRepo struct {
	rabbitv1.UnimplementedAlertServer
	calls int
	err   error
}

func (s *stubRabbitAlertRepo) ReceivePrometheusWebhook(_ context.Context, _ *rabbitv1.ReceivePrometheusWebhookRequest) (*rabbitv1.ReceivePrometheusWebhookReply, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return &rabbitv1.ReceivePrometheusWebhookReply{}, nil
}

func newTestRabbitAlertPusher(repo *stubRabbitAlertRepo, dedup *stubAlertPushDedup) *RabbitAlertPusher {
	return NewRabbitAlertPusher(repo, dedup, klog.NewHelper(klog.DefaultLogger))
}

func TestRabbitAlertPusherSkipsDuplicateFiringWithinDedupWindow(t *testing.T) {
	dedup := &stubAlertPushDedup{exists: true}
	repo := &stubRabbitAlertRepo{}
	pusher := newTestRabbitAlertPusher(repo, dedup)
	event := &bo.AlertEventBo{
		NamespaceUID: 1,
		Fingerprint:  "fp-1",
	}
	pusher.PushFiringAlert(context.Background(), 100, event)
	if repo.calls != 0 {
		t.Fatalf("expected no rabbit push, got %d calls", repo.calls)
	}
	if dedup.marked {
		t.Fatal("dedup mark should not run when push is skipped")
	}
}

func TestRabbitAlertPusherMarksDedupAfterSuccessfulFiringPush(t *testing.T) {
	dedup := &stubAlertPushDedup{}
	repo := &stubRabbitAlertRepo{}
	pusher := newTestRabbitAlertPusher(repo, dedup)
	event := &bo.AlertEventBo{
		NamespaceUID:     1,
		StrategyGroupUID: 2,
		StrategyUID:      3,
		LevelUID:         4,
		DatasourceUID:    5,
		StrategyName:     "HighCPU",
		LevelName:        "critical",
		Fingerprint:      "fp-1",
	}
	pusher.PushFiringAlert(context.Background(), 100, event)
	if repo.calls != 1 {
		t.Fatalf("expected one rabbit push, got %d calls", repo.calls)
	}
	if !dedup.marked {
		t.Fatal("expected dedup mark after successful push")
	}
}

func TestRabbitAlertPusherDoesNotMarkDedupWhenPushFails(t *testing.T) {
	dedup := &stubAlertPushDedup{}
	repo := &stubRabbitAlertRepo{err: errors.New("push failed")}
	pusher := newTestRabbitAlertPusher(repo, dedup)
	event := &bo.AlertEventBo{
		NamespaceUID: 1,
		Fingerprint:  "fp-1",
	}
	pusher.PushFiringAlert(context.Background(), 100, event)
	if dedup.marked {
		t.Fatal("dedup mark should not run when push fails")
	}
}

func TestRabbitAlertPusherCleanupDedupOnRecover(t *testing.T) {
	dedup := &stubAlertPushDedup{}
	repo := &stubRabbitAlertRepo{}
	pusher := newTestRabbitAlertPusher(repo, dedup)
	item := &bo.AlertEventItemBo{
		UID:          100,
		NamespaceUID: 1,
		Fingerprint:  "fp-1",
		StrategyName: "HighCPU",
		LevelName:    "critical",
	}
	if err := pusher.PushRecoveredAlert(context.Background(), item); err != nil {
		t.Fatalf("PushRecoveredAlert() error = %v", err)
	}
	if !dedup.cleared {
		t.Fatal("expected firing dedup key to be cleared on recover push")
	}
	if repo.calls != 1 {
		t.Fatalf("expected one recover push, got %d calls", repo.calls)
	}
}
