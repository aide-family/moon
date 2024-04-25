package consts

type TopicType string

const (
	AlertHookTopic        TopicType = "alert-hook"
	AgentOnlineTopic      TopicType = "online"
	ServerOnlineTopic     TopicType = "server-online"
	AgentOfflineTopic     TopicType = "offline"
	ServerOfflineTopic    TopicType = "server-offline"
	StrategyGroupAllTopic TopicType = "strategy-group-all"
	RemoveGroupTopic      TopicType = "rm-group-id"
)

// String 返回TopicType对应的字符串
func (t TopicType) String() string {
	switch t {
	case AlertHookTopic:
		return "[alert-hook] 边缘服务推送告警数据专用主题"
	case AgentOnlineTopic:
		return "[agent-online] 边缘节点在线状态推送主题"
	case AgentOfflineTopic:
		return "[agent-offline] 边缘节点离线状态推送主题"
	case StrategyGroupAllTopic:
		return "[strategy-group-all] 策略组所有节点数据推送主题"
	case RemoveGroupTopic:
		return "[remove-group] 策略组删除主题"
	case ServerOnlineTopic:
		return "[server-online] 服务器在线状态推送主题"
	case ServerOfflineTopic:
		return "[server-offline] 服务器离线状态推送主题"
	default:
		return "[" + string(t) + "] 未知主题, 管理员请注册或者拦截"
	}
}

// IsRegistered 判断TopicType是否已注册
func (t TopicType) IsRegistered() bool {
	switch t {
	case AlertHookTopic, AgentOnlineTopic, StrategyGroupAllTopic, AgentOfflineTopic, RemoveGroupTopic:
		return true
	default:
		return false
	}
}

// IsStrategyGroupAllTopic 判断TopicType是否为策略组所有节点数据推送主题
func (t TopicType) IsStrategyGroupAllTopic() bool {
	return t == StrategyGroupAllTopic
}
