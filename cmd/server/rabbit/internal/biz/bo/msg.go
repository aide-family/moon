package bo

// SendMsgParams 发送消息请求参数
type SendMsgParams struct {
	Route     string
	Data      []byte
	RequestID string
}

func (s *SendMsgParams) Key(app string) string {
	return "rabbit:" + app + ":" + s.Route + ":" + s.RequestID
}
