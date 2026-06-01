package feishu

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSignFeishuWebhookPayloadPreservesPostContent(t *testing.T) {
	raw := []byte(`{
  "msg_type": "post",
  "content": {
    "post": {
      "zh_cn": {
        "title": "alert",
        "content": [
          [
            {"tag": "text", "text": "status"},
            {"tag": "text", "text": "firing"}
          ]
        ]
      }
    }
  }
}`)

	signed, err := signFeishuWebhookPayload(raw, "1717248000", "secret")
	if err != nil {
		t.Fatalf("signFeishuWebhookPayload failed: %v", err)
	}
	if strings.Contains(string(signed), `"un_escape"`) {
		t.Fatalf("signed payload should not inject empty paragraph fields: %s", signed)
	}
	if !strings.Contains(string(signed), `"timestamp":"1717248000"`) {
		t.Fatalf("signed payload should include timestamp: %s", signed)
	}
	if !strings.Contains(string(signed), `"sign":`) {
		t.Fatalf("signed payload should include sign: %s", signed)
	}
}

func TestMessagePostRoundTripOmitsEmptyParagraphFields(t *testing.T) {
	raw := `{"msg_type":"post","content":{"post":{"zh_cn":{"title":"alert","content":[[{"tag":"text","text":"a"}]]}}}}`
	var msg Message
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	out, err := json.Marshal(&msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if strings.Contains(string(out), `"un_escape"`) {
		t.Fatalf("round-trip should omit empty paragraph fields: %s", out)
	}
}
