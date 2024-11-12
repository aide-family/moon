package alarmmodel

// Models 注册biz alarm model下相关模型
func Models() []any {
	return []any{
		&RealtimeAlarm{},
		&AlarmHistory{},
		&HistoryDetails{},
		&RealtimeDetails{},
		&AlarmRaw{},
		&RealtimeAlarmReceiver{},
		&RealtimeAlarmPage{},
		&AlarmSendHistory{},
	}
}
