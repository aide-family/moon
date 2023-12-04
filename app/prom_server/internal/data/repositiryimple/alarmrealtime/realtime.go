package alarmrealtime

import (
	"context"
	"strconv"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/consts"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/model/alarmscopes"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.AlarmRealtimeRepo = (*alarmRealtimeImpl)(nil)

type alarmRealtimeImpl struct {
	repository.UnimplementedAlarmRealtimeRepo
	log *log.Helper
	d   *data.Data

	query.IAction[model.PromAlarmRealtime]
}

func (l *alarmRealtimeImpl) Create(ctx context.Context, req ...*bo.AlarmRealtimeBO) ([]*bo.AlarmRealtimeBO, error) {
	newRealtimeModels := slices.To(req, func(req *bo.AlarmRealtimeBO) *model.PromAlarmRealtime { return req.ToModel() })
	if err := l.WithContext(ctx).Scopes(alarmscopes.ClauseOnConflict()).BatchCreate(newRealtimeModels, 50); err != nil {
		return nil, err
	}
	return slices.To(newRealtimeModels, func(m *model.PromAlarmRealtime) *bo.AlarmRealtimeBO { return bo.AlarmRealtimeModelToBO(m) }), nil
}

func (l *alarmRealtimeImpl) CacheByHistoryId(ctx context.Context, req ...*bo.AlarmRealtimeBO) error {
	// 写入redis hash表中
	args := make([]any, 0, len(req))
	for _, alarmRealtimeBO := range req {
		key := alarmRealtimeBO.ID
		args = append(args, key, alarmRealtimeBO)
	}

	key := consts.AlarmRealtimeCacheByHistoryId.String()
	return l.d.Client().HMSet(ctx, key, args...).Err()
}

func (l *alarmRealtimeImpl) DeleteCacheByHistoryId(ctx context.Context, historyId ...uint32) error {
	key := consts.AlarmRealtimeCacheByHistoryId.String()
	fields := slices.To(historyId, func(id uint32) string { return strconv.Itoa(int(id)) })
	return l.d.Client().HDel(ctx, key, fields...).Err()
}

func (l *alarmRealtimeImpl) AppendAlarmBeenNotifyMembers(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmBeenNotifyMemberBO) error {
	first, err := l.WithContext(ctx).FirstByID(uint(realtimeAlarmID))
	if err != nil {
		return err
	}

	return l.DB().WithContext(ctx).Model(first).Association(alarmscopes.RealtimeAssociationBeenNotifyMembers).Append(req.ToModel())
}

func (l *alarmRealtimeImpl) AppendAlarmBeenNotifyChatGroups(ctx context.Context, realtimeAlarmID uint32, req *bo.PromAlarmBeenNotifyChatGroupBO) error {
	first, err := l.WithContext(ctx).FirstByID(uint(realtimeAlarmID))
	if err != nil {
		return err
	}

	return l.DB().WithContext(ctx).Model(first).Association(alarmscopes.RealtimeAssociationBeenChatGroups).Append(req.ToModel())
}

func (l *alarmRealtimeImpl) GetRealtimeDetailById(ctx context.Context, id uint32, scopes ...query.ScopeMethod) (*bo.AlarmRealtimeBO, error) {
	first, err := l.WithContext(ctx).First(append(scopes, alarmscopes.InIds(id))...)
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
	return l.DB().WithContext(ctx).Model(first).Association(alarmscopes.RealtimeAssociationReplaceIntervenes).Append(newAlarmIntervene)
}

func (l *alarmRealtimeImpl) AlarmUpgrade(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmUpgradeBO) error {
	first, err := l.WithContext(ctx).FirstByID(uint(realtimeAlarmID))
	if err != nil {
		return err
	}
	newAlarmUpgrade := req.ToModel()
	return l.DB().WithContext(ctx).Model(first).Association(alarmscopes.RealtimeAssociationUpgradeInfo).Append(newAlarmUpgrade)
}

func (l *alarmRealtimeImpl) AlarmSuppress(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmSuppressBO) error {
	first, err := l.WithContext(ctx).FirstByID(uint(realtimeAlarmID))
	if err != nil {
		return err
	}
	newAlarmSuppress := req.ToModel()
	return l.DB().WithContext(ctx).Model(first).Association(alarmscopes.RealtimeAssociationSuppressInfo).Append(newAlarmSuppress)
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
