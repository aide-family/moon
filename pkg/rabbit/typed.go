package rabbit

type Message struct {
	ID string
	// 消息分组，同一个组的消息适用同一个消息抑制器
	Group string
	// Templates 表示该条消息需要使用哪些模版进行解析
	Templates []int64
	// Content 消息的内容
	Content any
}

type Template struct {
	ID int64
	// SuppressRuleID 抑制规则ID
	SuppressRuleID int64
	// Template 消息模版
	Template any
	//
	Secret []byte
}

type SuppressRule struct {
	ID       int64
	Type     string
	Interval int64
	Windows  int64
}
