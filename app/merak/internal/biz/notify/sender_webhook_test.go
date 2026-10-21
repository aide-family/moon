package notify

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/aide-family/Alioth/app/merak/pkg/bo"
)

func webhookTask(endpoint, app, secret string) bo.TaskSnapshot {
	return bo.TaskSnapshot{
		ChannelType: "WEBHOOK", ChannelApp: app, ChannelEndpoint: endpoint,
		ChannelSecret: secret, TemplateType: "WEBHOOK",
		Subject: "CPU 告警", Body: "cpu > 90%",
	}
}

func TestDingTalkSignatureAndPayload(t *testing.T) {
	const secret = "ding-secret"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timestamp := r.URL.Query().Get("timestamp")
		sign := r.URL.Query().Get("sign")
		parsed, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			t.Errorf("bad timestamp %q", timestamp)
		}
		if sign != dingTalkSign(secret, parsed) {
			t.Errorf("sign mismatch: %s", sign)
		}
		var payload struct {
			MsgType string `json:"msgtype"`
			Text    struct {
				Content string `json:"content"`
			} `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode: %v", err)
		}
		if payload.MsgType != "text" || payload.Text.Content != "CPU 告警\ncpu > 90%" {
			t.Errorf("payload = %+v", payload)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	result := Deliver(context.Background(), webhookTask(server.URL, "DINGTALK", secret))
	if !result.OK {
		t.Fatalf("deliver failed: %s", result.Detail)
	}
}

func TestFeishuSignatureAndPayload(t *testing.T) {
	const secret = "feishu-secret"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode: %v", err)
		}
		timestamp := payload["timestamp"].(string)
		sign := payload["sign"].(string)
		parsed, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			t.Errorf("bad timestamp %q", timestamp)
		}
		if sign != feishuSign(secret, parsed) {
			t.Errorf("sign mismatch: %s", sign)
		}
		content := payload["content"].(map[string]any)
		if content["text"] != "CPU 告警\ncpu > 90%" {
			t.Errorf("content = %v", content)
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"success"}`))
	}))
	defer server.Close()

	result := Deliver(context.Background(), webhookTask(server.URL, "FEISHU", secret))
	if !result.OK {
		t.Fatalf("deliver failed: %s", result.Detail)
	}
}

func TestWechatPayloadAndError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			MsgType string `json:"msgtype"`
			Text    struct {
				Content string `json:"content"`
			} `json:"text"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if payload.MsgType != "text" || payload.Text.Content == "" {
			t.Errorf("payload = %+v", payload)
		}
		_, _ = w.Write([]byte(`{"errcode":93000,"errmsg":"invalid webhook"}`))
	}))
	defer server.Close()

	result := Deliver(context.Background(), webhookTask(server.URL, "WECHAT", ""))
	if result.OK || !strings.Contains(result.Detail, "93000") {
		t.Fatalf("ok=%v detail=%s", result.OK, result.Detail)
	}
}

func TestOtherSignatureAndNoSecretInBody(t *testing.T) {
	const secret = "other-secret"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		timestamp := r.Header.Get("X-Alioth-Timestamp")
		signature := r.Header.Get("X-Alioth-Signature")
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(timestamp + "." + string(raw)))
		expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		if signature != expected {
			t.Errorf("signature = %s, want %s", signature, expected)
		}
		if strings.Contains(string(raw), secret) {
			t.Errorf("secret leaked in body: %s", raw)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	result := Deliver(context.Background(), webhookTask(server.URL, "OTHER", secret))
	if !result.OK {
		t.Fatalf("deliver failed: %s", result.Detail)
	}
}

func TestRenderRejectsBrokenTemplate(t *testing.T) {
	if _, err := render("{{.summary", `{"summary":"x"}`); err == nil {
		t.Fatal("broken template was accepted")
	}
	value, err := render("{{.summary}}", `{"summary":"cpu"}`)
	if err != nil || value != "cpu" {
		t.Fatalf("value=%q err=%v", value, err)
	}
}

func TestHeaderValueRemovesNewlines(t *testing.T) {
	if got := headerValue("a\r\nBcc: x@example.com"); strings.ContainsAny(got, "\r\n") {
		t.Fatalf("header still contains newline: %q", got)
	}
}

func TestFeishuSignUsesEmptyMessage(t *testing.T) {
	// 固定向量：key = "timestamp\nsecret"，message 为空。
	const secret = "s"
	const timestamp = int64(1700000000)
	mac := hmac.New(sha256.New, []byte("1700000000\ns"))
	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if got := feishuSign(secret, timestamp); got != want {
		t.Fatalf("feishuSign = %s, want %s", got, want)
	}
}
