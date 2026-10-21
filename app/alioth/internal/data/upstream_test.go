package data

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func TestCallUpstreamReportsRejectionReason(t *testing.T) {
	const reason = "unsupported URL format for path \"/api/v1/query\".\n" +
		"Make sure you're using cluster URL format"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(reason))
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	_, err := repo.InstantQuery(context.Background(), bo.Datasource{
		UID: "ds-vm", Driver: driverMetricsVictoriaMetrics, Endpoint: srv.URL,
	}, "up", 0)
	if err == nil {
		t.Fatal("expected an upstream error")
	}
	// 上游 400 的原因是配置或查询写错，属于调用方可修正的错误。
	if kerrors.Code(err) != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", kerrors.Code(err))
	}
	message := err.Error()
	if !strings.Contains(message, "unsupported URL format") || !strings.Contains(message, "cluster URL format") {
		t.Fatalf("message = %q", message)
	}
	if strings.ContainsAny(message, "\r\n") {
		t.Fatalf("message must be a single line, got %q", message)
	}
}

func TestCallUpstreamFallsBackToStatusText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	_, err := repo.InstantQuery(context.Background(), bo.Datasource{
		UID: "ds-vm", Driver: driverMetricsVictoriaMetrics, Endpoint: srv.URL,
	}, "up", 0)
	if err == nil {
		t.Fatal("expected an upstream error")
	}
	if !strings.Contains(err.Error(), http.StatusText(http.StatusBadGateway)) {
		t.Fatalf("message = %q", err.Error())
	}
}

func TestUpstreamReasonTruncatesLongBodies(t *testing.T) {
	reason := upstreamReason([]byte(strings.Repeat("错", maxUpstreamReason+50)))
	runes := []rune(reason)
	if len(runes) != maxUpstreamReason+1 || runes[len(runes)-1] != '…' {
		t.Fatalf("reason length = %d, tail = %q", len(runes), string(runes[len(runes)-1]))
	}
	if got := upstreamReason([]byte(" \n\t ")); got != "" {
		t.Fatalf("blank body reason = %q", got)
	}
}
