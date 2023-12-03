package alarmrealtime

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/model/alarm"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.AlarmRealtimeRepo = (*alarmRealtimeImpl)(nil)

type alarmRealtimeImpl struct {
	repository.UnimplementedAlarmRealtimeRepo
	log *log.Helper
	d   *data.Data

	query.IAction[model.PromAlarmRealtime]
}

func (l *alarmRealtimeImpl) AppendAlarmBeenNotifyMembers(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmBeenNotifyMemberBO) error {
	first, err := l.WithContext(ctx).FirstByID(uint(realtimeAlarmID))
	if err != nil {
		return err
	}

	return l.DB().WithContext(ctx).Model(first).Association(alarm.RealtimeAssociationBeenNotifyMembers).Append(req.ToModel())
}

func (l *alarmRealtimeImpl) AppendAlarmBeenNotifyChatGroups(ctx context.Context, realtimeAlarmID uint32, req *bo.PromAlarmBeenNotifyChatGroupBO) error {
	first, err := l.WithContext(ctx).FirstByID(uint(realtimeAlarmID))
	if err != nil {
		return err
	}

	return l.DB().WithContext(ctx).Model(first).Association(alarm.RealtimeAssociationBeenChatGroups).Append(req.ToModel())
}

func (l *alarmRealtimeImpl) GetRealtimeDetailById(ctx context.Context, id uint32, scopes ...query.ScopeMethod) (*bo.AlarmRealtimeBO, error) {
	first, err := l.WithContext(ctx).First(append(scopes, alarm.InIds(id))...)
	if err != nil {
		return nil, err
	}

	return bo.AlarmRealtimeModelToBO(first), nil
}

func (l *alarmRealtimeImpl) GetRealtimeList(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.AlarmRealtimeBO, error) {
	list, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	return slices.To(list, func(info *model.PromAlarmRealtime) *bo.AlarmRealtimeBO {
		return bo.AlarmRealtimeModelToBO(info)
	}), nil
}

func (l *alarmRealtimeImpl) AlarmIntervene(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmInterveneBO) error {
	first, err := l.WithContext(ctx).FirstByID(uint(realtimeAlarmID))
	if err != nil {
		return err
	}
	newAlarmIntervene := req.ToModel()
	return l.DB().WithContext(ctx).Model(first).Association(alarm.RealtimeAssociationReplaceIntervenes).Append(newAlarmIntervene)
}

func (l *alarmRealtimeImpl) AlarmUpgrade(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmUpgradeBO) error {
	first, err := l.WithContext(ctx).FirstByID(uint(realtimeAlarmID))
	if err != nil {
		return err
	}
	newAlarmUpgrade := req.ToModel()
	return l.DB().WithContext(ctx).Model(first).Association(alarm.RealtimeAssociationUpgradeInfo).Append(newAlarmUpgrade)
}

func (l *alarmRealtimeImpl) AlarmSuppress(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmSuppressBO) error {
	first, err := l.WithContext(ctx).FirstByID(uint(realtimeAlarmID))
	if err != nil {
		return err
	}
	newAlarmSuppress := req.ToModel()
	return l.DB().WithContext(ctx).Model(first).Association(alarm.RealtimeAssociationSuppressInfo).Append(newAlarmSuppress)
}

func (l *alarmRealtimeImpl) GetRealtimeCount(ctx context.Context, scopes ...query.ScopeMethod) (int64, error) {
	return l.WithContext(ctx).Count(scopes...)
}

func NewAlarmRealtime(d *data.Data, logger log.Logger) repository.AlarmRealtimeRepo {
	return &alarmRealtimeImpl{
		log: log.NewHelper(log.With(logger, "module", "repository.alarm.realtime")),
		d:   d,
		IAction: query.NewAction[model.PromAlarmRealtime](
			query.WithDB[model.PromAlarmRealtime](d.DB()),
		),
	}
}
