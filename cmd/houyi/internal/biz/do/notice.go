package do

type LabelNotices struct {
	Key            string   `json:"key,omitempty"`
	Value          string   `json:"value,omitempty"`
	ReceiverRoutes []string `json:"receiverRoutes,omitempty"`
}

func (l *LabelNotices) GetKey() string {
	if l == nil {
		return ""
	}
	return l.Key
}

func (l *LabelNotices) GetValue() string {
	if l == nil {
		return ""
	}
	return l.Value
}

func (l *LabelNotices) GetReceiverRoutes() []string {
	if l == nil {
		return nil
	}
	return l.ReceiverRoutes
}
