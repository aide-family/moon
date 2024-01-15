package strategy

type (
	AlarmStatus string

	Alarm struct {
		// 接收者
		Receiver string `json:"receiver"`
		// 报警状态
		Status AlarmStatus `json:"status"`
		// 告警列表
		Alerts []*Alarm `json:"alerts"`
		// 告警组标签
		Labels Labels `json:"labels"`
		// 公共标签
		CommonLabels map[string]string `json:"commonLabels"`
		// 公共注解
		CommonAnnotations map[string]string `json:"commonAnnotations"`
		// 外部链接
		ExternalURL string `json:"externalURL"`
		// 版本
		Version string `json:"version"`
		// 告警组key
		GroupKey string `json:"groupKey"`
		// 截断告警数
		TruncatedAlerts int32 `json:"truncate"`
	}

	Alert struct {
		// 告警状态
		Status AlarmStatus `json:"status"`
		// 标签
		Labels Labels `json:"labels"`
		// 注解
		Annotations Annotations `json:"annotations"`
		// 开始时间
		StartAt string `json:"startAt"`
		// 结束时间, 如果为空, 则表示未结束
		EndAt string `json:"endAt"`
		// 告警生成链接
		GeneratorURL string `json:"generatorURL"`
		// 指纹
		Fingerprint string `json:"fingerprint"`
	}
)
