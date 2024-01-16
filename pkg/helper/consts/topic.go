package consts

type TopicType string

const (
	AlertHookTopic   TopicType = "alert-hook"
	AgentOnlineTopic TopicType = "online"
)

// String 返回TopicType对应的字符串
func (t TopicType) String() string {
	switch t {
	case AlertHookTopic:
		return "[alert-hook] 边缘服务推送告警数据专用主题"
	case AgentOnlineTopic:
		return "[agent-online] 边缘节点在线状态推送主题"
	default:
		return "[" + string(t) + "] 未知主题, 管理员请注册或者拦截"
	}
}

// IsRegistered 判断TopicType是否已注册
func (t TopicType) IsRegistered() bool {
	switch t {
	case AlertHookTopic, AgentOnlineTopic:
		return true
	default:
		return false
	}
}
