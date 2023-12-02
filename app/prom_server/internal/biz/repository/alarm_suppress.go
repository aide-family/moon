package repository

var _ AlarmSuppressRepo = (*UnimplementedAlarmSuppressRepo)(nil)

type (
	AlarmSuppressRepo interface {
		unimplementedAlarmSuppressRepo()
	}

	UnimplementedAlarmSuppressRepo struct{}
)

func (UnimplementedAlarmSuppressRepo) unimplementedAlarmSuppressRepo() {}
