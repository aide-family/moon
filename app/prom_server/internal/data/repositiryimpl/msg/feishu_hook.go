package msg

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"prometheus-manager/pkg/httpx"
)

var _ HookNotify = (*feishuNotify)(nil)

type feishuNotify struct{}

func (l *feishuNotify) Alarm(url string, msg *HookNotifyMsg) error {
	notifyMsg := make(map[string]any)
	_ = json.Unmarshal([]byte(msg.Content), &notifyMsg)
	timeNow := time.Now()
	notifyMsg["timestamp"] = strconv.FormatInt(timeNow.Unix(), 10)
	notifyMsg["sign"] = GenSign(msg.Secret, timeNow.Unix())
	notifyMsgBytes, _ := json.Marshal(notifyMsg)
	_, err := httpx.NewHttpX().POST(url, notifyMsgBytes)
	return err
}

func NewFeishuNotify() HookNotify {
	return &feishuNotify{}
}

func GenSign(secret string, timestamp int64) string {
	//timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret

	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return ""
	}

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature
}
