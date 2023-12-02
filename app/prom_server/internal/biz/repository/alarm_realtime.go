package repository

var _ AlarmRealtimeRepo = (*UnimplementedAlarmRealtimeRepo)(nil)

type (
	AlarmRealtimeRepo interface {
		unimplementedAlarmRealtimeRepo()
	}

	UnimplementedAlarmRealtimeRepo struct{}
)

func (UnimplementedAlarmRealtimeRepo) unimplementedAlarmRealtimeRepo() {}
