package data

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent/probetargetgroup"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// TestProbeTargetGroupRequestEncrypted 目标组请求 JSON 以 enc:v1: 密文入库，
// 接口读取时解密成明文；历史明文行仍可读，重新保存后自动加密。
// 依赖 ALIOTH_TEST_DSN，未配置时跳过。
func TestProbeTargetGroupRequestEncrypted(t *testing.T) {
	store, ctx := newProbeFixStore(t)
	namespace := "ns-probe-cipher-" + time.Now().Format("150405.000000")
	t.Cleanup(cleanupProbeFixNamespace(store, namespace))
	repo := NewProbeRepo(store, slog.Default())

	plain := `{"timeout_ms":1500,"headers":{"Authorization":"Bearer plain-token"}}`
	group, err := repo.CreateProbeTargetGroup(ctx, bo.ProbeTargetGroup{
		NamespaceUID: namespace, Name: "加密目标组", Driver: "PROBE_HTTP",
		Labels: "{}", ExecutorLabels: "{}", Request: plain,
		Targets: []bo.ProbeTarget{{Spec: "https://example.com/health", Status: "ACTIVE"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	// 仓储返回值是解密后的明文，页面可编辑。
	if group.Request != plain {
		t.Fatalf("created request = %q, want plaintext", group.Request)
	}
	row, err := store.client.ProbeTargetGroup.Query().Where(probetargetgroup.UIDEQ(group.UID)).Only(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(row.Request, probeRequestCipherPrefix) {
		t.Fatalf("stored request should be encrypted, got %q", row.Request)
	}
	if strings.Contains(row.Request, "plain-token") {
		t.Fatalf("stored request should not contain plaintext token, got %q", row.Request)
	}

	// 历史明文行：直接写回明文，读取仍是明文；再保存时自动加密。
	if _, err := store.client.ProbeTargetGroup.Update().
		Where(probetargetgroup.UIDEQ(group.UID)).
		SetRequest(plain).Save(ctx); err != nil {
		t.Fatal(err)
	}
	legacy, err := repo.GetProbeTargetGroup(ctx, namespace, group.UID)
	if err != nil {
		t.Fatal(err)
	}
	if legacy.Request != plain {
		t.Fatalf("legacy request = %q, want plaintext", legacy.Request)
	}
	if _, err := repo.UpdateProbeTargetGroup(ctx, bo.ProbeTargetGroup{
		UID: group.UID, NamespaceUID: namespace, Name: "加密目标组", Driver: "PROBE_HTTP",
		Labels: "{}", ExecutorLabels: "{}", Request: legacy.Request,
		Targets: legacy.Targets,
	}); err != nil {
		t.Fatal(err)
	}
	reRow, err := store.client.ProbeTargetGroup.Query().Where(probetargetgroup.UIDEQ(group.UID)).Only(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(reRow.Request, probeRequestCipherPrefix) {
		t.Fatalf("request should be re-encrypted after save, got %q", reRow.Request)
	}
}

// TestProbeResultsM27ColumnsPersisted ICMP / DNS / TLS 追加列落库后能按窗口
// 读回，判定与页面展示都依赖这些值。依赖 ALIOTH_TEST_DSN，未配置时跳过。
func TestProbeResultsM27ColumnsPersisted(t *testing.T) {
	store, ctx := newProbeFixStore(t)
	namespace := "ns-probe-columns-" + time.Now().Format("150405.000000")
	t.Cleanup(cleanupProbeFixNamespace(store, namespace))
	repo := NewProbeRepo(store, slog.Default())

	group, err := repo.CreateProbeTargetGroup(ctx, bo.ProbeTargetGroup{
		NamespaceUID: namespace, Name: "结果列目标组", Driver: "PROBE_ICMP",
		Labels: "{}", ExecutorLabels: "{}", Request: "{}",
		Targets: []bo.ProbeTarget{{Spec: "127.0.0.1", Status: "ACTIVE"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(group.Targets) != 1 {
		t.Fatalf("created targets = %d, want 1", len(group.Targets))
	}
	targetUID := group.Targets[0].UID

	slot := time.Now().UTC().Truncate(time.Second)
	if _, err := repo.SaveProbeResults(ctx, []bo.ProbeResult{{
		NamespaceUID: namespace, GroupUID: group.UID, TargetUID: targetUID,
		Slot: slot, ProbedAt: slot, OK: true, Value: "0", DurationMS: 12,
		PacketLoss: 40, JitterMS: 7, MinRTTMS: 12, MaxRTTMS: 33,
		AnswerCount: 3, DNSAnswers: "1.1.1.1, 2.2.2.2",
		TLSVersion: "TLSv1.3", TLSCipher: "TLS_AES_128_GCM_SHA256",
		CertSubject: "CN=example.com", CertIssuer: "CN=Test CA", CertSANS: "example.com",
		Executor: "manual",
	}}); err != nil {
		t.Fatal(err)
	}

	results, err := repo.ListProbeResults(ctx, namespace, []string{targetUID}, slot.Add(-time.Minute), slot.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("listed results = %d, want 1", len(results))
	}
	got := results[0]
	if got.PacketLoss != 40 || got.JitterMS != 7 || got.MinRTTMS != 12 || got.MaxRTTMS != 33 {
		t.Fatalf("icmp columns = loss:%d jitter:%d min:%d max:%d", got.PacketLoss, got.JitterMS, got.MinRTTMS, got.MaxRTTMS)
	}
	if got.AnswerCount != 3 || got.DNSAnswers != "1.1.1.1, 2.2.2.2" {
		t.Fatalf("dns columns = count:%d answers:%q", got.AnswerCount, got.DNSAnswers)
	}
	if got.TLSVersion != "TLSv1.3" || got.TLSCipher != "TLS_AES_128_GCM_SHA256" {
		t.Fatalf("tls columns = version:%q cipher:%q", got.TLSVersion, got.TLSCipher)
	}
	if got.CertSubject != "CN=example.com" || got.CertIssuer != "CN=Test CA" || got.CertSANS != "example.com" {
		t.Fatalf("cert columns = subject:%q issuer:%q sans:%q", got.CertSubject, got.CertIssuer, got.CertSANS)
	}
}
