package data

import (
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/pkg/merr"
)

// newProbeTestBox 构造测试用加密盒，密钥派生与 Dependencies 使用同一份实现。
func newProbeTestBox(t *testing.T, key string) *secretBox {
	t.Helper()
	box, err := newSecretBox(key)
	if err != nil {
		t.Fatalf("newSecretBox: %v", err)
	}
	return box
}

// TestProbeRequestCipherRoundTrip 请求 JSON 加密后带 enc:v1: 前缀且不含明文
// token，解密后与原文一致；对已加密值再次加密不会叠加前缀。
func TestProbeRequestCipherRoundTrip(t *testing.T) {
	box := newProbeTestBox(t, "dev-only-encrypt-key")
	plain := `{"headers":{"Authorization":"Bearer token"},"timeout_ms":1500}`

	cipher, err := encryptProbeRequest(box, plain)
	if err != nil {
		t.Fatalf("encryptProbeRequest: %v", err)
	}
	if !strings.HasPrefix(cipher, probeRequestCipherPrefix) {
		t.Fatalf("cipher %q should have prefix %q", cipher, probeRequestCipherPrefix)
	}
	if strings.Contains(cipher, "Bearer") || strings.Contains(cipher, "timeout_ms") {
		t.Fatalf("cipher should not contain plaintext request, got %q", cipher)
	}

	again, err := encryptProbeRequest(box, cipher)
	if err != nil {
		t.Fatalf("re-encrypt: %v", err)
	}
	if again != cipher {
		t.Fatalf("already encrypted request should be kept, got %q", again)
	}

	decoded, err := decryptProbeRequest(box, cipher)
	if err != nil {
		t.Fatalf("decryptProbeRequest: %v", err)
	}
	if decoded != plain {
		t.Fatalf("decoded = %q, want %q", decoded, plain)
	}
}

// TestProbeRequestCipherLegacyPlaintext 没有前缀的历史明文行按原样读取，下次
// 保存时由加密路径自动升级；空请求保持空串。
func TestProbeRequestCipherLegacyPlaintext(t *testing.T) {
	box := newProbeTestBox(t, "dev-only-encrypt-key")
	legacy := `{"method":"GET","expect_status":"200-299","timeout_ms":3000}`
	decoded, err := decryptProbeRequest(box, legacy)
	if err != nil {
		t.Fatalf("decrypt legacy plaintext: %v", err)
	}
	if decoded != legacy {
		t.Fatalf("legacy plaintext = %q, want %q", decoded, legacy)
	}

	if empty, err := encryptProbeRequest(box, ""); err != nil || empty != "" {
		t.Fatalf("encrypt empty = %q, err = %v, want empty", empty, err)
	}
	if empty, err := decryptProbeRequest(box, ""); err != nil || empty != "" {
		t.Fatalf("decrypt empty = %q, err = %v, want empty", empty, err)
	}
}

// TestProbeRequestCipherWrongKey 密钥不一致时解密失败要返回服务端错误，不能
// 静默当成空请求，否则执行器会用默认参数探测错误的对象。
func TestProbeRequestCipherWrongKey(t *testing.T) {
	box := newProbeTestBox(t, "key-a")
	other := newProbeTestBox(t, "key-b")
	cipher, err := encryptProbeRequest(box, `{"timeout_ms":3000}`)
	if err != nil {
		t.Fatalf("encryptProbeRequest: %v", err)
	}
	if _, err := decryptProbeRequest(other, cipher); err == nil {
		t.Fatal("decrypt with a different key should fail")
	} else if merr.IsInvalid(err) {
		t.Fatalf("decrypt failure should be a server error, got client error: %v", err)
	}
}

// TestProbeGroupDecryptsRequest 仓储读取目标组时解密请求，历史明文行仍返回
// 明文，接口与试探测看到的是可直接执行的请求参数。
func TestProbeGroupDecryptsRequest(t *testing.T) {
	box := newProbeTestBox(t, "dev-only-encrypt-key")
	repo := &probeRepo{&repoBase{box: box}}

	plain := `{"timeout_ms":1500,"dns_record_type":"AAAA"}`
	cipher, err := encryptProbeRequest(box, plain)
	if err != nil {
		t.Fatalf("encryptProbeRequest: %v", err)
	}
	item, err := repo.toProbeGroup(&ent.ProbeTargetGroup{UID: "group-cipher", Request: cipher})
	if err != nil {
		t.Fatalf("toProbeGroup: %v", err)
	}
	if item.Request != plain {
		t.Fatalf("decrypted request = %q, want %q", item.Request, plain)
	}

	legacy, err := repo.toProbeGroup(&ent.ProbeTargetGroup{UID: "group-legacy", Request: plain})
	if err != nil {
		t.Fatalf("toProbeGroup legacy: %v", err)
	}
	if legacy.Request != plain {
		t.Fatalf("legacy request = %q, want %q", legacy.Request, plain)
	}
}
