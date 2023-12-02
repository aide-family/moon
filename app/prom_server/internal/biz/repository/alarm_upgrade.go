package repository

var _ AlarmUpgradeRepo = (*UnimplementedAlarmUpgradeRepo)(nil)

type (
	AlarmUpgradeRepo interface {
		unimplementedAlarmUpgradeRepo()
	}

	UnimplementedAlarmUpgradeRepo struct{}
)

func (UnimplementedAlarmUpgradeRepo) unimplementedAlarmUpgradeRepo() {}
