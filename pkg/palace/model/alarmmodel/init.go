package alarmmodel

// AlarmModels 注册biz alarm model下相关模型
func AlarmModels() []any {
	return []any{
		&RealtimeAlarm{},
		&AlarmHistory{},
		&HistoryFields{},
		&RealtimeFields{},
	}
}
