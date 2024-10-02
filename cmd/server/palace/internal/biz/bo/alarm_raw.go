package bo

type (

	// CreateAlarmRawParams 创建告警原始数据参数
	CreateAlarmRawParams struct {
		Fingerprint string `json:"fingerprint"`
		RawInfo     string `json:"rawInfo"`
		TeamID      uint32 `json:"teamId"`
	}
)
