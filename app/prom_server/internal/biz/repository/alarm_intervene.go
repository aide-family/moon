package repository

var _ AlarmInterveneRepo = (*UnimplementedAlarmInterveneRepo)(nil)

type (
	AlarmInterveneRepo interface {
		unimplementedInterveneRepo()
	}

	UnimplementedAlarmInterveneRepo struct{}
)

func (UnimplementedAlarmInterveneRepo) unimplementedInterveneRepo() {}
